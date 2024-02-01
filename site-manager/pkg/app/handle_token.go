package app

import (
	"context"
	"fmt"
	"strings"

	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	kube_config "github.com/netcracker/drnavigator/site-manager/config/kube_config"
	"github.com/netcracker/drnavigator/site-manager/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	toolsWatch "k8s.io/client-go/tools/watch"
)

const smServiceAccountName = "sm-auth-sa"

// handleToken gets token of sm-auth-sa from kubernetes when it updates
func HandleToken(tokenChannel chan string, errChannel chan error) {
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
