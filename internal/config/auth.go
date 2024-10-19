package config

import "golang-mongodb-rest-api-starter/internal/env"

type AuthConfig struct {
	AccessSecret  string
	RefreshSecret string
}

func LoadAuthConfig() AuthConfig {
	return AuthConfig{
		AccessSecret:  env.GetEnvWithDefault("ACCESS_SECRET", "TODO-CHANGE-THIS"),
		RefreshSecret: env.GetEnvWithDefault("REFRESH_SECRET", "TODO-CHANGE-THIS"),
	}
}
