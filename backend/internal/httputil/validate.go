package httputil

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Validate checks the struct against its `validate` tags.
// Returns nil if valid, or a descriptive error listing field violations.
func Validate(v any) error {
	if err := validate.Struct(v); err != nil {
		var errs validator.ValidationErrors
		if ok := isValidationErrors(err, &errs); ok {
			msgs := make([]string, 0, len(errs))
			for _, fe := range errs {
				msgs = append(msgs, fieldError(fe))
			}
			return fmt.Errorf("%s", strings.Join(msgs, "; "))
		}
		return err
	}
	return nil
}

// WriteValidationError writes a 422 JSON response with the validation error message.
func WriteValidationError(w http.ResponseWriter, err error) {
	WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
}

func isValidationErrors(err error, out *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*out = ve
		return true
	}
	return false
}

func fieldError(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s failed validation: %s", field, fe.Tag())
	}
}
