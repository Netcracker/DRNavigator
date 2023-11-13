package logger

import (
	envconfig "github.com/netcracker/drnavigator/site-manager-cr-controller/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger = nil

func getZapLogLevel() (zap.AtomicLevel, error) {
	if envconfig.EnvConfig.Debug {
		return zap.NewAtomicLevelAt(zapcore.DebugLevel), nil
	}
	return zap.NewAtomicLevelAt(zapcore.InfoLevel), nil
}

// SimpleLogger returns app loger or initializes the new one, if it's not defined yet
func SimpleLogger() *zap.SugaredLogger {
	if log == nil {
		logLevel, err := getZapLogLevel()
		if err != nil {
			panic(err)
		}

		cfg := zap.Config{
			Level:            logLevel,
			Encoding:         "console",
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:        "T",
				MessageKey:     "M",
				LevelKey:       "L",
				NameKey:        "N",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.CapitalLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
		}

		logger, err := cfg.Build()
		if err != nil {
			panic(err)
		}

		log = logger.Sugar()
	}
	return log
}
