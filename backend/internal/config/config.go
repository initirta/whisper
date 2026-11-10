package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port           string
	DatabaseURL    string
	FrontendOrigin string
	SecureCookies  bool
}

func Load() (Config, error) {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = strings.TrimSpace(os.Getenv("BACKEND_PORT"))
	}
	if port == "" {
		port = "8080"
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required; load the local .env file before starting the API")
	}
	frontendOrigin := strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN")), "/")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:3000"
	}
	secureCookies := false
	if value := strings.TrimSpace(os.Getenv("SESSION_COOKIE_SECURE")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, errors.New("SESSION_COOKIE_SECURE must be true or false")
		}
		secureCookies = parsed
	}

	return Config{Port: port, DatabaseURL: databaseURL, FrontendOrigin: frontendOrigin, SecureCookies: secureCookies}, nil
}
