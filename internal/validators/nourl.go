package validators

import (
	"golang-mongodb-rest-api-starter/internal/utils/strutils"
	"strings"

	"github.com/go-playground/validator/v10"
)

type NoURL struct{}

func (n NoURL) Validate(fl validator.FieldLevel) bool {
	// Get the string value of the field
	fieldValue := fl.Field().String()

	// Split the string into words
	words := strings.Fields(fieldValue)

	// Iterate through words and check for the presence a url
	for _, word := range words {
		if strutils.StringContainsUrl(word) {
			return false
		}
	}
	return true
}
