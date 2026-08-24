package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type appConfig struct {
	Environment        string
	DatabaseURL        string
	FrontendOrigin     string
	Port               string
	StaticDir          string
	AtCoderProblemsURL string
}

func loadConfig() (appConfig, error) {
	config := appConfig{Environment: valueOr("APP_ENV", "development"), DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")), FrontendOrigin: valueOr("FRONTEND_ORIGIN", "http://localhost:3000"), Port: valueOr("PORT", "8080"), StaticDir: strings.TrimSpace(os.Getenv("STATIC_DIR")), AtCoderProblemsURL: valueOr("ATCODER_PROBLEMS_URL", "https://kenkoooo.com/atcoder")}
	port, err := strconv.Atoi(config.Port)
	if err != nil || port < 1 || port > 65535 {
		return appConfig{}, fmt.Errorf("PORT must be between 1 and 65535")
	}
	if config.DatabaseURL == "" {
		return appConfig{}, fmt.Errorf("DATABASE_URL is required")
	}
	for name, value := range map[string]string{"FRONTEND_ORIGIN": config.FrontendOrigin, "ATCODER_PROBLEMS_URL": config.AtCoderProblemsURL} {
		parsed, err := url.Parse(value)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return appConfig{}, fmt.Errorf("%s must be an absolute HTTP URL", name)
		}
	}
	return config, nil
}
func valueOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
