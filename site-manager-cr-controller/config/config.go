package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug          bool   `envconfig:"SM_DEBUG" default:"false"`
	KubeconfigFile string `envconfig:"SM_KUBECONFIG_FILE"`

	CRGroup   string `envconfig:"SM_GROUP" default:"netcracker.com"`
	CRPrural  string `envconfig:"SM_PRURAL" default:"sitemanagers"`
	CRVersion string `envconfig:"SM_VERSION" default:"v3"`

	PostRequestTimeout int64 `envconfig:"SM_POST_REQUEST_TIMEOUT" default:"30"`
}

var EnvConfig Config

// InitConfig initializes the configuration from the environment variables
func InitConfig() error {
	return envconfig.Process("SM", &EnvConfig)
}
