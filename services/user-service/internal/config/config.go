package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresDSN string
	RedisAddr   string
	GRPCPort    string
}

func Load() *Config {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := &Config{
		PostgresDSN: os.Getenv("POSTGRES_DSN"),
		RedisAddr:   os.Getenv("REDIS_ADDR"),
		GRPCPort:    os.Getenv("GRPC_PORT"),
	}
	fmt.Println("cfg", cfg)

	if cfg.PostgresDSN == "" {
		log.Fatal("POSTGRES_DSN not set")
	}

	return cfg
}
