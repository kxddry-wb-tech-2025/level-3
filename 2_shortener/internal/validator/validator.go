package validator

import "github.com/go-playground/validator/v10"

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{validate: validator.New()}
}

func (v *Validator) URL(url string) error {
	return v.validate.Var(url, "required,url")
}

func (v *Validator) Struct(s interface{}) error {
	return v.validate.Struct(s)
}
