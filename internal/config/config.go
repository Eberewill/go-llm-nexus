package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	LLM      LLMConfig
}

type ServerConfig struct {
	Port string `mapstructure:"SERVER_PORT"`
	Env  string `mapstructure:"ENV"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"REDIS_ADDR"`
	Password string `mapstructure:"REDIS_PASSWORD"`
}

type LLMConfig struct {
	OpenAIKey             string  `mapstructure:"OPENAI_API_KEY"`
	GeminiKey             string  `mapstructure:"GEMINI_API_KEY"`
	OpenAIModel           string  `mapstructure:"OPENAI_MODEL"`
	GeminiModel           string  `mapstructure:"GEMINI_MODEL"`
	OpenAIInputCostPer1K  float64 `mapstructure:"OPENAI_INPUT_COST_PER_1K"`
	OpenAIOutputCostPer1K float64 `mapstructure:"OPENAI_OUTPUT_COST_PER_1K"`
	GeminiInputCostPer1K  float64 `mapstructure:"GEMINI_INPUT_COST_PER_1K"`
	GeminiOutputCostPer1K float64 `mapstructure:"GEMINI_OUTPUT_COST_PER_1K"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Replace dots with underscores in env variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("OPENAI_MODEL", "gpt-3.5-turbo")
	viper.SetDefault("GEMINI_MODEL", "gemini-2.0-flash-exp")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; ignore error if desired or rely on env vars
	}

	keys := []string{
		"SERVER_PORT",
		"ENV",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"REDIS_ADDR",
		"REDIS_PASSWORD",
		"OPENAI_API_KEY",
		"GEMINI_API_KEY",
		"OPENAI_MODEL",
		"GEMINI_MODEL",
		"OPENAI_INPUT_COST_PER_1K",
		"OPENAI_OUTPUT_COST_PER_1K",
		"GEMINI_INPUT_COST_PER_1K",
		"GEMINI_OUTPUT_COST_PER_1K",
	}
	for _, key := range keys {
		if err := viper.BindEnv(key); err != nil {
			return nil, fmt.Errorf("failed to bind env var %s: %w", key, err)
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
			Env:  viper.GetString("ENV"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
		},
		Redis: RedisConfig{
			Addr:     viper.GetString("REDIS_ADDR"),
			Password: viper.GetString("REDIS_PASSWORD"),
		},
		LLM: LLMConfig{
			OpenAIKey:             viper.GetString("OPENAI_API_KEY"),
			GeminiKey:             viper.GetString("GEMINI_API_KEY"),
			OpenAIModel:           viper.GetString("OPENAI_MODEL"),
			GeminiModel:           viper.GetString("GEMINI_MODEL"),
			OpenAIInputCostPer1K:  viper.GetFloat64("OPENAI_INPUT_COST_PER_1K"),
			OpenAIOutputCostPer1K: viper.GetFloat64("OPENAI_OUTPUT_COST_PER_1K"),
			GeminiInputCostPer1K:  viper.GetFloat64("GEMINI_INPUT_COST_PER_1K"),
			GeminiOutputCostPer1K: viper.GetFloat64("GEMINI_OUTPUT_COST_PER_1K"),
		},
	}

	return cfg, nil
}
