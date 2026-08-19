package goninja

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validatorInstance is shared across all Validate calls; the tag name func
// makes ValidationError.Fields use JSON field names instead of Go field
// names, matching what the client actually sent.
var validatorInstance = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})
	return v
}

// Validate runs go-playground/validator's struct validation (driven by the
// `validate` struct tags on a generated Create/Update schema) and, on
// failure, returns a ValidationError keyed by JSON field name rather than
// the raw validator.ValidationErrors — so generated handlers and
// DefaultErrorMapper don't need to know about the underlying library.
func Validate(v any) error {
	if err := validatorInstance.Struct(v); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			fields := make(map[string]string, len(verrs))
			for _, fe := range verrs {
				fields[fe.Field()] = fe.Tag()
			}
			return ValidationError{Fields: fields}
		}
		return err
	}
	return nil
}
