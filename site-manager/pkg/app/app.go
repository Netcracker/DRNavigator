package app

import (
	"context"
	"fmt"
	"strings"

	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	kube_config "github.com/netcracker/drnavigator/site-manager/config/kube_config"
	"github.com/netcracker/drnavigator/site-manager/logger"
	cr_client "github.com/netcracker/drnavigator/site-manager/pkg/client/cr"
	"github.com/netcracker/drnavigator/site-manager/pkg/controllers"
	"github.com/netcracker/drnavigator/site-manager/pkg/model"
	"github.com/netcracker/drnavigator/site-manager/pkg/service"
	"github.com/netcracker/drnavigator/site-manager/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	toolsWatch "k8s.io/client-go/tools/watch"
	ctrl "sigs.k8s.io/controller-runtime"
)

const smServiceAccountName = "sm-auth-sa"

var appLog = ctrl.Log.WithName("app")

// Serve starts new https server app
func Serve(bindAddress string, bindWebhookAddress string, bindMetricsAddress string, devMode bool, certDir string, certFile string, keyFile string) error {
	// Collects sm config
	smConfig := &model.SMConfig{TokenChannel: make(chan string)}
	if smConfigFile := envconfig.EnvConfig.SMConfigFile; smConfigFile != "" {
		appLog.V(1).Info("SMConfig file detected", smConfigFile)
		if err := utils.ParseYamlFile(smConfigFile, smConfig); err != nil {
			return fmt.Errorf("error parsing sm config file: %s", err)
		}
	}

	// Initialize services
	logger.SetupLogger()
	mgr, crManager, err := initializeServices(smConfig, bindWebhookAddress, bindMetricsAddress, devMode, certDir, certFile, keyFile)
	if err != nil {
		return err
	}

	// initialize cross gorutine error, that is used for every worked gorutine until it returns an error
	errorChannel := make(chan error)

	// Handle token if authorization is enabled in separate gorutine
	if !smConfig.Testing.Enabled && (envconfig.EnvConfig.BackHttpAuth || envconfig.EnvConfig.FrontHttpAuth) {
		go handleToken(smConfig.TokenChannel, errorChannel)
	}

	// initialize api for webhook in separate gorutine
	if bindWebhookAddress != "" {
		go ServeWebhookServer(mgr, crManager, errorChannel)
	}
	// initialize api for main site-manager in separate gorutine
	go ServeMainServer(bindAddress, bindWebhookAddress, certDir, certFile, keyFile, crManager, smConfig, errorChannel)

	// wait when some gorutine returns an error
	return <-errorChannel
}

// initializeServices initializes used services depended from configuration
func initializeServices(smConfig *model.SMConfig, bindWebhookAddress string, bindMetricsAddress string, devMode bool, certDir string, certFile string, keyFile string) (ctrl.Manager, service.CRManager, error) {
	if !smConfig.Testing.Enabled {
		// init controller-runtime manager
		mgr, err := controllers.NewControllerRuntimeManager(bindWebhookAddress, bindMetricsAddress, devMode, certDir, certFile, keyFile)
		if err != nil {
			return nil, nil, fmt.Errorf("error initializing controller runtime manager: %s", err)
		}
		// init cr client
		crClient, err := cr_client.NewNewCRClient(mgr.GetClient())
		if err != nil {
			return nil, nil, fmt.Errorf("error initializing cr-client: %s", err)
		}
		controllers.SetupCRReconciler(crClient, mgr)
		crManager, err := service.NewCRManager(smConfig, crClient)
		if err != nil {
			return nil, nil, fmt.Errorf("error initializing cr manager service: %s", err)
		}
		return mgr, crManager, nil
	} else {
		crManager, err := service.NewCRManager(smConfig, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("error initializing cr manager service: %s", err)
		}
		return nil, crManager, nil
	}
}

// handleToken gets token of sm-auth-sa from kubernetes when it updates
func handleToken(tokenChannel chan string, errChannel chan error) {
	config, err := kube_config.GetKubeConfig()
	if err != nil {
		errChannel <- fmt.Errorf("can't create kubeconfig to handle token from secret %s", err)
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		errChannel <- fmt.Errorf("can't initialize kubernetes client to handle token from secret %s", err)
		return
	}

	namespace := envconfig.EnvConfig.PodNamespace
	appLog.Info("Namespace detected", "namespace", namespace)

	timeout := int64(30)
	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		return clientset.CoreV1().ServiceAccounts(namespace).Watch(context.Background(), metav1.ListOptions{TimeoutSeconds: &timeout})
	}

	watcher, err := toolsWatch.NewRetryWatcher("1", &cache.ListWatch{WatchFunc: watchFunc})
	if err != nil {
		errChannel <- fmt.Errorf("can't initialize watcher to handle token from secret %s", err)
		return
	}

	var token string
	for {
		select {
		case tokenChannel <- token:
		case event, ok := <-watcher.ResultChan():
			if !ok {
				appLog.Error(nil, "Watch SA event channel is closed")
				errChannel <- fmt.Errorf("can't handle token: channel is closed")
				return
			}
			serviceAccount, ok := event.Object.(*corev1.ServiceAccount)
			if !ok {
				appLog.Error(nil, "Can't get SA from event")
				errChannel <- fmt.Errorf("can't handle SA from watching event")
				return
			}
			if serviceAccount.GetName() != smServiceAccountName {
				continue
			}
			if event.Type == watch.Added || event.Type == watch.Modified {
				secretRef := utils.FindFirstFromSlice(serviceAccount.Secrets, func(secretRef corev1.ObjectReference) bool {
					return strings.Contains(secretRef.Name, "token")
				})
				if secretRef == nil {
					appLog.Info("Secret for appropriate SA is not ready yet")
					continue
				}
				secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), secretRef.Name, metav1.GetOptions{})
				if err != nil {
					appLog.Error(err, "Can't get secret", "secret-name", secretRef.Name, "namespace", namespace)
					errChannel <- fmt.Errorf("can't handle token from secret %s for SA %s", secretRef.Name, smServiceAccountName)
					return
				}
				if btoken, found := secret.Data["token"]; !found {
					appLog.Error(nil, "Can't get token from secret", "secret-name", secretRef.Name, "namespace", namespace)
					errChannel <- fmt.Errorf("can't handle token from secret %s for SA %s", secretRef.Name, smServiceAccountName)
					return
				} else {
					token = string(btoken)
				}
				appLog.V(1).Info("Service-account event. Token was updated.", "sa-name", smServiceAccountName, "event-type", event.Type)
			} else if event.Type == watch.Deleted {
				appLog.Error(nil, "Service-account was deleted. Exit", "sa-name", smServiceAccountName)
				errChannel <- fmt.Errorf("service-account %s was deleted", smServiceAccountName)
				return
			}
		}
	}
}
