package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	HTTPPort  string
	MasterDSN string
	SlaveDSNs []string
	RedisAddr string
	RedisPassword string
	RedisDB   int
}

const DefaultHTTPPort = ":8080"

func NewConfig() (*Config, error) {
	cfg := Config{}

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		cfg.HTTPPort = DefaultHTTPPort
	} else {
		// Ensure port starts with ':' if not already present
		if len(httpPort) > 0 && httpPort[0] != ':' {
			cfg.HTTPPort = ":" + httpPort
		} else {
			cfg.HTTPPort = httpPort
		}
	}
	masterDSN := os.Getenv("MASTER_DSN")
	if masterDSN != "" {
		cfg.MasterDSN = masterDSN
	}

	// Redis configuration
	cfg.RedisAddr = os.Getenv("REDIS_ADDR")
	if cfg.RedisAddr == "" {
		cfg.RedisAddr = "localhost:6379"
	}
	cfg.RedisPassword = os.Getenv("REDIS_PASSWORD")
	cfg.RedisDB = 0 // default DB

	return &cfg, nil
}
