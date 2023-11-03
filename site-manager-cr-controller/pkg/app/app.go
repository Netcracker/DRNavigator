package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/labstack/gommon/log"
	envconfig "github.com/netcracker/drnavigator/site-manager-cr-controller/config"
	kube_config "github.com/netcracker/drnavigator/site-manager-cr-controller/config/kube_config"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/logger"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/model"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/service"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	toolsWatch "k8s.io/client-go/tools/watch"
)

const smServiceAccountName = "sm-auth-sa"

// Serve starts new https server app
func Serve(bindAddress string, bindWebhookAddress string, certFile string, keyFile string) error {
	// Collects sm config
	smConfig := &model.SMConfig{TokenChannel: make(chan string)}
	if smConfigFile := envconfig.EnvConfig.SMConfigFile; smConfigFile != "" {
		log.Debug("SMConfig file detected: %s", smConfigFile)
		if err := utils.ParseYamlFile(smConfigFile, smConfig); err != nil {
			return fmt.Errorf("error parsing sm config file: %s", err)
		}
	}

	// Initalize services
	crManager, err := service.NewCRManager(smConfig)
	if err != nil {
		return fmt.Errorf("can't initialize cr manager service: %s", err)
	}
	crValidator, err := service.NewValidator(smConfig)
	if err != nil {
		return fmt.Errorf("can't initialize cr validator: %s", err)
	}
	crConverter, err := service.NewConverter(smConfig)
	if err != nil {
		return fmt.Errorf("can't initialize cr converter: %s", err)
	}

	// initialize cross gorutine error, that is used for every worked gorutine until it returns an error
	errorChannel := make(chan error)

	// Handle token if authorization is enabled in separate gorutine
	if !smConfig.Testing.Enabled && (envconfig.EnvConfig.BackHttpAuth || envconfig.EnvConfig.FrontHttpAuth) {
		go handleToken(smConfig.TokenChannel, errorChannel)
	}

	// initialize api for webhook in separate gorutine
	if bindWebhookAddress != "" {
		go ServeWebhookServer(bindWebhookAddress, certFile, keyFile, crValidator, crConverter, errorChannel)
	}
	// initialize api for main site-manager in separate gorutine
	go ServeWMainServer(bindAddress, bindWebhookAddress, certFile, keyFile, crManager, smConfig, errorChannel)

	// wait when some gorutine returns an error
	return <-errorChannel
}

// handleToken gets token of sm-auth-sa from kubernetes when it updates
func handleToken(TokenChannel chan string, errChannel chan error) {
	logger := logger.SimpleLogger()
	config, err := kube_config.GetKubeConfig()
	if err != nil {
		errChannel <- fmt.Errorf("Can't create kubeconfig to handle token from secret %s", err)
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		errChannel <- fmt.Errorf("Can't initialize kubernetes client to handle token from secret %s", err)
		return
	}

	namespace := envconfig.EnvConfig.PodNamespace
	logger.Infof("Current namespace: %s", namespace)

	timeout := int64(30)
	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		return clientset.CoreV1().ServiceAccounts(namespace).Watch(context.Background(), metav1.ListOptions{TimeoutSeconds: &timeout})
	}

	watcher, err := toolsWatch.NewRetryWatcher("1", &cache.ListWatch{WatchFunc: watchFunc})
	if err != nil {
		errChannel <- fmt.Errorf("Can't initialize watcher to handle token from secret %s", err)
		return
	}

	for {
		event, ok := <-watcher.ResultChan()
		if !ok {
			logger.Errorf("Watch SA event channel is closed")
			errChannel <- fmt.Errorf("can't handle token: channel is closed")
		}
		serviceAccount := event.Object.(*corev1.ServiceAccount)
		if serviceAccount.GetName() != smServiceAccountName {
			continue
		}
		if event.Type == watch.Added || event.Type == watch.Modified {
			secretRef := utils.FindFirstFromSlice(serviceAccount.Secrets, func(secretRef corev1.ObjectReference) bool {
				return strings.Contains(secretRef.Name, "token")
			})
			if secretRef == nil {
				logger.Warnf("Secret for appropriate SA is not ready yet")
				continue
			}
			secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), secretRef.Name, metav1.GetOptions{})
			if err != nil {
				logger.Errorf("Can't get secret with name %s from namespace %s: %s", secretRef.Name, namespace, err)
				errChannel <- fmt.Errorf("can't handle token from secret %s for SA %s", secretRef.Name, smServiceAccountName)
				continue
			}
			btoken, found := secret.Data["token"]
			if !found {
				logger.Errorf("Can't get token from secret with name %s from namespace %s: %s", secretRef.Name, namespace, err)
				errChannel <- fmt.Errorf("can't handle token from secret %s for SA %s", secretRef.Name, smServiceAccountName)
				continue
			}
			TokenChannel <- string(btoken)
			logger.Infof("Service-account %s was %s. Token was updated.", smServiceAccountName, event.Type)
		} else if event.Type == watch.Deleted {
			logger.Errorf("Service-account %s was deleted. Exit", smServiceAccountName)
			errChannel <- fmt.Errorf("service-account %s was deleted", smServiceAccountName)
		}
	}
}
