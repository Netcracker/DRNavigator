package model

// SMConfig represents the debug configuration, that is used in SM_CONFIG_FILE
type SMConfig struct {
	Testing      SMConfigTesting `yaml:"testing"`
}

// SMConfigTesting represents the testing field from SMConfig
type SMConfigTesting struct {
	Enabled bool         `yaml:"enabled"`
	Token   string       `yaml:"token"`
	SMDict  SMDictionary `yaml:"sm_dict"`
}
