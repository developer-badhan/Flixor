package validator

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

/**
 * Package validator wraps go-playground/validator/v10 and converts its error
 * output into the structured FieldError slice that our response envelope expects.
 * Why wrap it?
 *   - go-playground/validator returns error strings like "Key: 'LoginInput.Email'
 *     Error:Field validation for 'Email' failed on the 'email' tag" — ugly for clients.
 *   - We transform these into {"field":"email","message":"must be a valid email"} objects.
 *   - Centralised custom messages mean you change wording in one place.
*/

/**
 * Singleton validator instance
*/
var (
	instance *validator.Validate
	once     sync.Once
)

// get returns the singleton validator, initialising it on first call.
func get() *validator.Validate {
	once.Do(func() {
		instance = validator.New()

		// Register custom tag name function so field names in errors are
		// taken from the json struct tag (e.g. "email") not the Go field name ("Email").
		instance.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	})
	return instance
}

/**
 * Public API:
 * FieldError describes a single invalid field.
*/
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

/**
 * Validate validates a struct and returns a slice of FieldError.
 * Returns nil if validation passes.
 *
 * Usage:
 *
 * 	if errs := validator.Validate(input); errs != nil {
 * 	    _ = c.Error(apperror.ErrValidation.WithDetails(errs))
 * 	    return
 * 	}
*/
func Validate(s any) []FieldError {
	err := get().Struct(s)
	if err == nil {
		return nil
	}

	var fieldErrors []FieldError

	// go-playground/validator returns ValidationErrors for struct validation
	var ve validator.ValidationErrors
	if ok := isValidationErrors(err, &ve); ok {
		for _, fe := range ve {
			fieldErrors = append(fieldErrors, FieldError{
				Field:   fe.Field(),
				Message: humanMessage(fe),
			})
		}
		return fieldErrors
	}

	// Fallback for invalid input types
	fieldErrors = append(fieldErrors, FieldError{
		Field:   "input",
		Message: err.Error(),
	})
	return fieldErrors
}

/**
 * ValidateVar validates a single variable against a tag string.
 * Usage:
 * 	if errs := validator.ValidateVar(email, "required,email"); errs != nil { ... }
*/
func ValidateVar(field any, tag string) []FieldError {
	err := get().Var(field, tag)
	if err == nil {
		return nil
	}
	return []FieldError{{Field: "value", Message: err.Error()}}
}

/**
 * Private helpers
 * humanMessage converts a validator.FieldError into a friendly message.
 * Add more cases here as your project grows.
*/
func humanMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", fe.Field())
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", fe.Field())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s failed validation: %s", fe.Field(), fe.Tag())
	}
}

/**
 * isValidationErrors is a helper because errors.As doesn't work directly with
 * go-playground/validator's type.
*/
func isValidationErrors(err error, target *validator.ValidationErrors) bool {
	ve, ok := err.(validator.ValidationErrors)
	if ok {
		*target = ve
	}
	return ok
}
