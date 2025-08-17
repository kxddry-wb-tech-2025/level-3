package validator

import "github.com/go-playground/validator/v10"

// Validator is the validator for the requests.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator instance.
func New() *Validator {
	return &Validator{validate: validator.New()}
}

// URL validates the URL.
func (v *Validator) URL(url string) error {
	return v.validate.Var(url, "required,url")
}

// Struct validates the struct.
func (v *Validator) Struct(s interface{}) error {
	return v.validate.Struct(s)
}
