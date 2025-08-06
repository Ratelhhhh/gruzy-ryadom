package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Server   ServerConfig   `yaml:"server"`
	Bots     BotsConfig
	Env      string `yaml:"env"`
}

type DatabaseConfig struct {
	URL string `yaml:"url"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type BotsConfig struct {
	DriverBotToken string
	AdminBotToken  string
}

func Load() (*Config, error) {
	config := &Config{}
	
	// Try to read YAML config file (optional)
	configData, err := os.ReadFile("config.yaml")
	if err == nil {
		if err := yaml.Unmarshal(configData, config); err != nil {
			return nil, fmt.Errorf("failed to parse config.yaml: %w", err)
		}
	} else {
		// If config.yaml doesn't exist, use defaults
		config.Server.Port = "8080"
		config.Env = "development"
	}

	// Override with environment variables (environment variables take precedence)
	if dbURL := getEnv("DATABASE_URL", ""); dbURL != "" {
		config.Database.URL = dbURL
	}
	
	if port := getEnv("PORT", ""); port != "" {
		config.Server.Port = port
	}

	// Load bot tokens from environment variables (required)
	config.Bots = BotsConfig{
		DriverBotToken: getEnv("DRIVER_BOT_TOKEN", ""),
		AdminBotToken:  getEnv("ADMIN_BOT_TOKEN", ""),
	}

	// Validate required fields
	if config.Bots.DriverBotToken == "" {
		return nil, fmt.Errorf("DRIVER_BOT_TOKEN is required")
	}
	if config.Bots.AdminBotToken == "" {
		return nil, fmt.Errorf("ADMIN_BOT_TOKEN is required")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
