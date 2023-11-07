package model

// SMConfig represents the debug configuration, that is used in SM_CONFIG_FILE
type SMConfig struct {
	Testing      SMConfigTesting `yaml:"testing"`
	TokenChannel chan string     `yaml:"-"`
}

// SMConfigTesting represents the testing field from SMConfig
type SMConfigTesting struct {
	Enabled bool         `yaml:"enabled"`
	Token   string       `yaml:"token"`
	SMDict  SMDictionary `yaml:"sm_dict"`
}

// GetToken handles token changes from channel and return the last one
func (smc *SMConfig) GetToken() string {
	if smc.Testing.Enabled {
		return smc.Testing.Token
	}
	return <-smc.TokenChannel
}
