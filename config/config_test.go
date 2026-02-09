package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		cfg, err := NewConfig()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.HTTPPort != DefaultHTTPPort {
			t.Errorf("Expected default port '%s', got '%s'", DefaultHTTPPort, cfg.HTTPPort)
		}

		if cfg.RedisAddr != "localhost:6379" {
			t.Errorf("Expected default Redis addr 'localhost:6379', got '%s'", cfg.RedisAddr)
		}
	})

	t.Run("Custom HTTP port", func(t *testing.T) {
		os.Setenv("HTTP_PORT", "9090")
		defer os.Unsetenv("HTTP_PORT")

		cfg, err := NewConfig()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.HTTPPort != ":9090" {
			t.Errorf("Expected port ':9090', got '%s'", cfg.HTTPPort)
		}
	})

	t.Run("Custom HTTP port with colon", func(t *testing.T) {
		os.Setenv("HTTP_PORT", ":8081")
		defer os.Unsetenv("HTTP_PORT")

		cfg, err := NewConfig()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.HTTPPort != ":8081" {
			t.Errorf("Expected port ':8081', got '%s'", cfg.HTTPPort)
		}
	})

	t.Run("Custom Redis address", func(t *testing.T) {
		os.Setenv("REDIS_ADDR", "redis:6379")
		defer os.Unsetenv("REDIS_ADDR")

		cfg, err := NewConfig()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.RedisAddr != "redis:6379" {
			t.Errorf("Expected Redis addr 'redis:6379', got '%s'", cfg.RedisAddr)
		}
	})

	t.Run("Custom master DSN", func(t *testing.T) {
		testDSN := "postgres://test:test@localhost:5432/test"
		os.Setenv("MASTER_DSN", testDSN)
		defer os.Unsetenv("MASTER_DSN")

		cfg, err := NewConfig()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.MasterDSN != testDSN {
			t.Errorf("Expected DSN '%s', got '%s'", testDSN, cfg.MasterDSN)
		}
	})
}
