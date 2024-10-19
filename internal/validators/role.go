package validators

import (
	"golang-mongodb-rest-api-starter/internal/constants"

	"github.com/go-playground/validator/v10"
)

type Role struct{}

func (r Role) Validate(fl validator.FieldLevel) bool {
	// Get the string value of the field
	fieldValue := fl.Field().String()
	for _, role := range constants.GetAllowedRoles() {
		// fmt.Println(role)
		if role == fieldValue {
			return true
		}
	}
	return false
}
