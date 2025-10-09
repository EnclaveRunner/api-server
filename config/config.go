package config

import (
	enclaveConfig "github.com/EnclaveRunner/shareddeps/config"
)

type AppConfig struct {
	enclaveConfig.BaseConfig `mapstructure:",squash"`
}


var Cfg = &AppConfig{}
