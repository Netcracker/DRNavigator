package logger

import (
	"github.com/netcracker/drnavigator/site-manager-cr-controller/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger = nil

func GetZapLogLevel() (zap.AtomicLevel, error) {
	if config.EnvConfig.Debug {
		return zap.NewAtomicLevelAt(zapcore.DebugLevel), nil
	}
	return zap.NewAtomicLevelAt(zapcore.InfoLevel), nil
}

func SimpleLogger() *zap.SugaredLogger {
	if log == nil {
		logLevel, err := GetZapLogLevel()
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
		zap.NewDevelopmentEncoderConfig()

		logger, err := cfg.Build()
		if err != nil {
			panic(err)
		}

		log = logger.Sugar()
	}
	return log
}
