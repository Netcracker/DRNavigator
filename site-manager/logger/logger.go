package logger

import (
	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// getLoggerOptions calculates options for logger
func getLoggerOptions() *zap.Options {
	return &zap.Options{
		Development: envconfig.EnvConfig.Debug,
	}
}

// SetupLogger initialize new logger and regists it as controller-runtime logger
func SetupLogger() {
	opts := getLoggerOptions()
	log := zap.New(zap.UseFlagOptions(opts))
	ctrl.SetLogger(log)
}
