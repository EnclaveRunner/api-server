package config

import (
	enclaveConfig "github.com/EnclaveRunner/shareddeps/config"
)

type AppConfig struct {
	enclaveConfig.BaseConfig `mapstructure:",squash"`

	Database struct {
		Host     string `mapstructure:"host"     validate:"required,hostname|ip"`
		Port     int    `mapstructure:"port"     validate:"required,numeric,min=1,max=65535"`
		Username string `mapstructure:"username" validate:"required"`
		Password string `mapstructure:"password" validate:"required"`
		Database string `mapstructure:"database" validate:"required"`
		SSLMode  string `mapstructure:"sslmode"  validate:"oneof=disable require verify-ca verify-full"`
	} `mapstructure:"database" validate:"required"`

	Admin struct {
		Username    string `mapstructure:"username"     validate:"required"`
		DisplayName string `mapstructure:"display_name" validate:"required"`
		Password    string `mapstructure:"password"     validate:"required"`
	} `mapstructure:"admin" validate:"required"`

	ArtifactRegistry struct {
		Host string `mapstructure:"host" validate:"required,hostname|ip"`
		Port int    `mapstructure:"port" validate:"required,numeric,min=1,max=65535"`
	} `mapstructure:"artifact_registry" validate:"required"`
}

var Cfg = &AppConfig{}
