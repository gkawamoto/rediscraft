package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MinecraftFolder string `envconfig:"MINECRAFT_FOLDER" required:"true"`
	MinecraftMemory string `envconfig:"MINECRAFT_MEMORY" default:"2G"`
	MinecraftJAR    string `envconfig:"MINECRAFT_JAR" default:"minecraft_server.1.16.5.jar"`
	RedisAddr       string `envconfig:"REDIS_ADDR" default:":6379"`
	RedisPassword   string `envconfig:"REDIS_PASSWORD" required:"true"`
}

func New() (Config, error) {
	var config Config

	err := envconfig.Process("", &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
