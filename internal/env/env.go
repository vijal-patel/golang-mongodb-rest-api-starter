package env

import (
	"golang-mongodb-rest-api-starter/internal/constants"
	"os"
)

func GetEnvWithDefault(key string, defaultValue string) string {
	val := os.Getenv(key)
	if val == constants.EmptyString {
		val = defaultValue
	}
	return val
}
