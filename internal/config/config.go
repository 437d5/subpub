package config

import (
	"log"
	"os"

	"github.com/go-yaml/yaml"
)

type Config struct {
	GRPCPort string `yaml:"grpcPort"`
}

func Load() Config {
	var cfg Config

	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Printf("read yaml file failed: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		log.Printf("unmarshal config file failed: %v", err)
	}

	return cfg
}
