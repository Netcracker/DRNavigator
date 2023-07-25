package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug          bool   `envconfig:"SM_DEBUG" default:"false"`
	KubeconfigFile string `envconfig:"SM_KUBECONFIG_FILE"`

	CRGroup   string `envconfig:"SM_GROUP" default:"netcracker.com"`
	CRPrural  string `envconfig:"SM_PRURAL" default:"sitemanagers"`
	CRVersion string `envconfig:"SM_KIND" default:"v3"`
}

var EnvConfig Config

func InitConfig() error {
	return envconfig.Process("SM", &EnvConfig)
}
