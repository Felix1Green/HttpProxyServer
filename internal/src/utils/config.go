package utils

import (
	"github.com/spf13/viper"
	"log"
)

type ProxyConfig struct{
	Port string
	CertFilePath string
	KeyFilePath string
}


func GetConfig(configPath string) (*ProxyConfig, error) {
	if configPath == ""{
		return &ProxyConfig{
			":8000",
			"",
			"",
		}, nil
	}

	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("Unable to read config file: %s", err)
		return nil, err
	}

	config := new(ProxyConfig)
	config.Port = viper.GetString("Port")
	config.KeyFilePath = viper.GetString("KeyFile")
	config.CertFilePath = viper.GetString("CertFile")

	return config, nil
}