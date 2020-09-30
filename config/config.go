package config

import "time"

type StorageSettings struct {
	Host     string `json:"host",yaml:"host"`
	Port     int    `json:"port",yaml:"port"`
	Username string `json:"username",yaml:"username"`
	Password string `json:"password",yaml:"password"`
}

type Config struct {
	MasterSecret    string          `json:"masterSecret",yaml:"masterSecret"`
	Port            int             `json:"port",yaml:"port"`
	StorageType     string          `json:"storage",yaml:"storage"`
	StorageSettings StorageSettings `json:"storageSettings",yaml:"storageSettings"`
	WriteDelay      time.Duration   `json:"delay",yaml:"delay`
}

func DefaultConfig() *Config {
	return &Config{
		StorageType: "memory",
		Port:        8080,
		// Not really any simple config tools in stdLib, would use Viper here
		MasterSecret: "imarealtivelylongandsomewhatsecuresecret",
		WriteDelay:   5,
	}
}
