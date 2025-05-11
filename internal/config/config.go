package config

import (
	"log"
	"os"
)

const (
	GRPC_PORT_DEFAULT = "5000"
)

type Config struct {
	GRPCPort string
}

func Load() Config {
	var cfg Config

	cfg.GRPCPort = loadEnvOrDefault("GRPC_PORT", GRPC_PORT_DEFAULT)

	log.Printf("Config initialized: %v", cfg)
	return cfg
}

func loadEnvOrDefault(name, defaultVal string) string {
	value := os.Getenv(name)
	if value == "" {
		value = defaultVal
	}

	return value
}