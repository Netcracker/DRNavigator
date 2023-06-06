package logger

import (
	"encoding/json"
	"os"

	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetZapLogLevel() (zap.AtomicLevel, error) {
	if paas_debug_env, exist := os.LookupEnv("PAAS_DEBUG"); exist {
		debug_enabled, err := strconv.ParseBool(paas_debug_env)
		if err != nil {
			return zap.AtomicLevel{}, err
		}
		if debug_enabled {
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
	log_level, err := GetZapLogLevel()
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

	cfg.Level = log_level
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger.Sugar()
}
