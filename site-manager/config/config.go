package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// common env configuration
	Debug          bool   `envconfig:"SM_DEBUG" default:"false"`
	SMConfigFile   string `envconfig:"SM_CONFIG_FILE"`
	KubeconfigFile string `envconfig:"SM_KUBECONFIG_FILE"`
	PodNamespace   string `envconfig:"POD_NAMESPACE" default:"site-manager"`

	// http configuration
	HttpScheme    string `envconfig:"HTTP_SCHEME" default:"http://"`
	HttpsEnaled   bool   `envconfig:"HTTPS_ENABLED" default:"false"`
	FrontHttpAuth bool   `envconfig:"FRONT_HTTP_AUTH" default:"false"`
	BackHttpAuth  bool   `envconfig:"BACK_HTTP_AUTH" default:"false"`
	SMCaCert      string `envconfig:"SM_CACERT" default:"True"`

	// cr configuration
	CRGroup    string `envconfig:"SM_GROUP" default:"netcracker.com"`
	CRKind     string `envconfig:"SM_KIND" default:"SiteManager"`
	CRKindList string `envconfig:"SM_KIND_LIST" default:"SiteManagerList"`

	// timeouts
	PostRequestTimeout int64 `envconfig:"SM_POST_REQUEST_TIMEOUT" default:"30"`
	GetRequestTimeout  int64 `envconfig:"SM_GET_REQUEST_TIMEOUT" default:"10"`
}

var EnvConfig Config

// InitConfig initializes the configuration from the environment variables
func InitConfig() error {
	return envconfig.Process("SM", &EnvConfig)
}
