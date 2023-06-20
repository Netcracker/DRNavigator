package logger

import (
	"encoding/json"
	"os"

	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetZapLogLevel() (zap.AtomicLevel, error) {
	if paasDebugEnv, exist := os.LookupEnv("PAAS_DEBUG"); exist {
		debugEnabled, err := strconv.ParseBool(paasDebugEnv)
		if err != nil {
			return zap.AtomicLevel{}, err
		}
		if debugEnabled {
			return zap.NewAtomicLevelAt(zapcore.DebugLevel), nil
		} else {
			return zap.NewAtomicLevelAt(zapcore.InfoLevel), nil
		}
	} else {
		return zap.NewAtomicLevelAt(zapcore.InfoLevel), nil
	}
}

// SimpleLogger is used to create simple logger without any configuration.
func SimpleLogger() *zap.SugaredLogger {
	logLevel, err := GetZapLogLevel()
	if err != nil {
		panic(err)
	}
	rawJSON := []byte(`{
		"level": "info",
		"encoding": "console",
		"outputPaths": ["stdout"],
		"errorOutputPaths": ["stderr"],
		"encoderConfig": {
		  "messageKey": "message",
		  "levelKey": "level",
		  "nameKey": "name",
		  "consoleSeparator": "    "
		}
	  }`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}

	cfg.Level = logLevel
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger.Sugar()
}
