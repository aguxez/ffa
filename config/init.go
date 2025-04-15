package config

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Environment                    string `mapstructure:"environment"`
	DatabaseURL                    string `mapstructure:"database_url"`
	TelegramBotAPIKey              string `mapstructure:"telegram_bot_api_key"`
	OpenRouterAPIKey               string `mapstructure:"openrouter_api_key"`
	MealPlanFeedbackRequestMinutes int64  `mapstructure:"meal_plan_feedback_request_minutes"`
}

// NewAppConfig creates a new AppConfig instance from the "config.yaml" file.
// If the file does not exist, it will use defaults and environment variables.
// If the file exists but there is an error reading it, it will log an error and return nil.
// If there is an error unmarshalling the config, it will log an error and return nil.
// Otherwise it will return the AppConfig instance.
func NewAppConfig() (AppConfig, error) {
	envErr := godotenv.Load()

	// Setup Viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetDefault("environment", "development")

	// Read config file
	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info().Msg("Config file not found. Using defaults and environment variables")
		} else {
			log.Err(err).Msg("Error reading config file")
			return AppConfig{}, err
		}
	}

	err = v.BindEnv("environment")
	if err != nil {
		log.Err(err).Msg("Error binding environment variable")
		return AppConfig{}, err
	}
	err = v.BindEnv("database_url")
	if err != nil {
		log.Err(err).Msg("Error binding database_url variable")
		return AppConfig{}, err
	}
	err = v.BindEnv("telegram_bot_api_key")
	if err != nil {
		log.Err(err).Msg("Error binding telegram_bot_api_key variable")
		return AppConfig{}, err
	}

	err = v.BindEnv("openrouter_api_key")
	if err != nil {
		log.Err(err).Msg("Error binding openrouter_api_key variable")
		return AppConfig{}, err
	}
	err = v.BindEnv("meal_plan_feedback_request_minutes")
	if err != nil {
		log.Err(err).Msg("Error binding meal_plan_feedback_request_minutes variable")
		return AppConfig{}, err
	}

	// Unmarshal to struct
	var config AppConfig
	err = v.Unmarshal(&config)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshalling config")
		return AppConfig{}, err
	}

	if envErr != nil && config.Environment == "development" {
		log.
			Err(envErr).
			Msg("Error loading environment variables for non dev env")

		return AppConfig{}, envErr
	}

	return config, nil
}
