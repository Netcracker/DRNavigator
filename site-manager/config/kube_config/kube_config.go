package kube_config

import (
	"fmt"

	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	"github.com/netcracker/drnavigator/site-manager/pkg/utils"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetKubeConfig() (*rest.Config, error) {
	var config *rest.Config
	var err error

	if envconfig.EnvConfig.KubeconfigFile != "" {
		if err := utils.CheckFile(envconfig.EnvConfig.KubeconfigFile); err != nil {
			return nil, fmt.Errorf("error getting kubeconfig file: %s", err)
		}
		config, err = clientcmd.BuildConfigFromFlags("", envconfig.EnvConfig.KubeconfigFile)
		if err != nil {
			return nil, fmt.Errorf("error config for kubernetes client: %s", err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("error config for kubernetes client: %s", err)
		}
	}
	return config, nil
}
