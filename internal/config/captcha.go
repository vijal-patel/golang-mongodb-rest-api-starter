package config

import (
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/env"
	"log"
)

type CaptchaConfig struct {
	Secret     string
	CaptchaUrl string
}

func LoadCaptchaConfig() CaptchaConfig {
	secret := env.GetEnvWithDefault("CAPTCHA_SECRET", constants.EmptyString)
	url := env.GetEnvWithDefault("CAPTCHA_URL", constants.EmptyString)
	if secret == constants.EmptyString || url == constants.EmptyString {
		log.Fatal("Invalid captcha config")
	}
	return CaptchaConfig{
		Secret:     secret,
		CaptchaUrl: url,
	}
}
