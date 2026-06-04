package configs

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	JWTSecret             string
	AppBaseURL            string
	PayHereMerchantID     string
	PayHereMerchantSecret string
	PayHereCheckoutURL    string
	PayHereReturnURL      string
	PayHereCancelURL      string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		Port:                  env("PORT", "3000"),
		DatabaseURL:           os.Getenv("DB_URL"),
		JWTSecret:             env("JWT_SECRET", "change_me_in_development"),
		AppBaseURL:            env("APP_BASE_URL", "http://127.0.0.1:3000"),
		PayHereMerchantID:     os.Getenv("PAYHERE_MERCHANT_ID"),
		PayHereMerchantSecret: os.Getenv("PAYHERE_MERCHANT_SECRET"),
		PayHereCheckoutURL:    env("PAYHERE_CHECKOUT_URL", "https://sandbox.payhere.lk/pay/checkout"),
		PayHereReturnURL:      env("PAYHERE_RETURN_URL", "http://127.0.0.1:3000/payment/success"),
		PayHereCancelURL:      env("PAYHERE_CANCEL_URL", "http://127.0.0.1:3000/payment/cancel"),
	}, nil
}

func env(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	return value
}
