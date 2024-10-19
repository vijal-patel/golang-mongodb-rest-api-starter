package config

import (
	"golang-mongodb-rest-api-starter/internal/constants"
	"log"
	"os"
)

type EmailConfig struct {
	ApiKey       string
	ApiSecret    string
	FromName     string
	FromEmail    string
	ReplyToEmail string
}

func LoadEmailConfig() EmailConfig {
	apiKey := os.Getenv("EMAIL_API_KEY")
	apiSecret := os.Getenv("EMAIL_API_SECRET")
	fromName := os.Getenv("EMAIL_FROM_NAME")
	fromEmail := os.Getenv("EMAIL_FROM_EMAIL")
	replyToEmail := os.Getenv("EMAIL_REPLY_TO_EMAIL")

	if apiKey == constants.EmptyString || apiSecret == constants.EmptyString || fromName == constants.EmptyString || fromEmail == constants.EmptyString {
		log.Fatal("Invalid email config")
	}
	return EmailConfig{
		ApiKey:       apiKey,
		ApiSecret:    apiSecret,
		FromName:     fromName,
		FromEmail:    fromEmail,
		ReplyToEmail: replyToEmail,
	}
}
