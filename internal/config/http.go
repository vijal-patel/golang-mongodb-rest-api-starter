package config

import (
	"fmt"
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/env"
	"os"
	"strings"
)

type HTTPConfig struct {
	Host           string
	Port           string
	ExposePort     string
	AllowedOrigins []string
}

func LoadHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Host:           env.GetEnvWithDefault("HOST", constants.LocalHost),
		Port:           env.GetEnvWithDefault("PORT", "8080"),
		ExposePort:     os.Getenv("EXPOSE_PORT"),
		AllowedOrigins: strings.Split(env.GetEnvWithDefault("ALLOWED_ORIGINS", fmt.Sprintf("http://%s:3000", constants.LocalHost)), ","),
	}
}
