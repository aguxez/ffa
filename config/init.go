package config

import (
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Environment string `mapstructure:"environment"`
	Database    struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"database"`
}

// NewAppConfig creates a new AppConfig instance from the "config.yaml" file.
// If the file does not exist, it will use defaults and environment variables.
// If the file exists but there is an error reading it, it will log an error and return nil.
// If there is an error unmarshalling the config, it will log an error and return nil.
// Otherwise it will return the AppConfig instance.
func NewAppConfig() *AppConfig {
	// Setup Viper
	v := viper.New() // Create a new Viper instance
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetDefault("environment", "development")

	// Configure Viper to replace dots with underscores for env vars
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("")
	v.AutomaticEnv()

	// Read config file
	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info().Msg("Config file not found. Using defaults and environment variables")
		} else {
			log.Error().Err(err).Msg("Error reading config file")
			return nil
		}
	}

	// Override with environment variables if they exist
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		v.Set("database.url", dbURL)
	}

	// Unmarshal to struct
	var config AppConfig
	err = v.Unmarshal(&config)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshalling config")
		return nil
	}

	return &config
}
