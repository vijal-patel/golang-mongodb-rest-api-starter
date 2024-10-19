package validators

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func RegisterValidators(v *validator.Validate) {

	if err := v.RegisterValidation("nourl", NoURL{}.Validate); err != nil {
		fmt.Printf("Error registering norul validation rule: %s\n", err)
		return
	}

	if err := v.RegisterValidation("role", Role{}.Validate); err != nil {
		fmt.Printf("Error registering role validation rule: %s\n", err)
		return
	}
}
