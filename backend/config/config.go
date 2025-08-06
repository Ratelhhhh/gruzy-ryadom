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
	// Read YAML config file
	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config.yaml: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(configData, config); err != nil {
		return nil, fmt.Errorf("failed to parse config.yaml: %w", err)
	}

	// Load bot tokens from environment variables
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