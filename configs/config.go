package configs

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	StripeSecretKey     string
	StripeWebhookSecret string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		Port:                env("PORT", "3000"),
		DatabaseURL:         os.Getenv("DB_URL"),
		JWTSecret:           env("JWT_SECRET", "change_me_in_development"),
		StripeSecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
	}, nil
}

func env(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	return value
}
