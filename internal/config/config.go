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
	OpenAIKey string `mapstructure:"OPENAI_API_KEY"`
	GeminiKey string `mapstructure:"GEMINI_API_KEY"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Replace dots with underscores in env variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; ignore error if desired or rely on env vars
	}

	var cfg Config
	
	// Manually bind env vars if needed, or rely on mapstructure tags with AutomaticEnv
	// Viper's unmarshal with mapstructure tags works well.
	
	// We need to bind specific keys if the structure doesn't match env vars 1:1 automatically
	// But here we used mapstructure tags which Viper respects.
	
	// However, Viper AutomaticEnv doesn't automatically map ENV_VAR to Nested.Field
	// unless we set a prefix or bind them manually. 
	// A simpler approach for this demo is to just bind the struct.
	
	// Let's bind manually for clarity or use a flat config structure if preferred.
	// For nested, we usually do:
	viper.BindEnv("SERVER_PORT")
	viper.BindEnv("ENV")
	
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("DB_NAME")
	
	viper.BindEnv("REDIS_ADDR")
	viper.BindEnv("REDIS_PASSWORD")
	
	viper.BindEnv("OPENAI_API_KEY")
	viper.BindEnv("GEMINI_API_KEY")

	if err := viper.Unmarshal(&cfg.Server); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&cfg.Database); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&cfg.Redis); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&cfg.LLM); err != nil {
		return nil, err
	}

	return &cfg, nil
}
