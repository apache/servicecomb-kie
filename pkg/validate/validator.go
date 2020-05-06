package validate

import (
	"errors"
	"fmt"
	"strings"

	ut "github.com/go-playground/universal-translator"
	valid "github.com/go-playground/validator"
)

var errorTranslator ut.Translator // no use but as an index

func registerErrorTranslator(_ ut.Translator) error { return nil }

// Validator validates data
// not safe, use it after initialized
type Validator struct {
	rules map[string]*RegexValidateRule
	valid *valid.Validate
}

// Validate validates the input data
func (v *Validator) Validate(i interface{}) error {
	err := v.valid.Struct(i)
	if err != nil {
		return v.wrapError(err)
	}
	return nil
}

// converts the raw error into an easy-to-understand error
func (v *Validator) wrapError(err error) error {
	validErr, ok := err.(valid.ValidationErrors)
	if !ok {
		return err
	}
	msgs := make([]string, len(validErr))
	for i, ve := range validErr {
		fe := ve.(valid.FieldError)
		msgs[i] = fe.Translate(errorTranslator)
	}
	return errors.New("validate failed, " + strings.Join(msgs, " | "))
}

// RegisterRule registers a custom validate rule
func (v *Validator) RegisterRule(r *RegexValidateRule) error {
	if r == nil {
		return errors.New("empty regex validate rule")
	}
	v.rules[r.tag] = r
	if err := v.valid.RegisterValidation(r.tag, r.validateFL); err != nil {
		return err
	}
	return v.AddErrorTranslation4Tag(r.tag)
}

// translates raw errors to easy-to-understand messages
func (v *Validator) translateError(_ ut.Translator, fe valid.FieldError) string {
	var rule string
	if r, ok := v.rules[fe.Tag()]; ok {
		rule = r.Explain()
	} else {
		rule = fe.Tag()
		if len(fe.Param()) > 0 {
			rule = rule + " = " + fe.Param()
		}
	}
	return fmt.Sprintf("field: %s, rule: %s", fe.Namespace(), rule)
}

// AddErrorTranslation4Tag adds translation for the errors of some tag,
// to make the error easier to understand
func (v *Validator) AddErrorTranslation4Tag(tag string) error {
	return v.valid.RegisterTranslation(tag,
		errorTranslator,
		registerErrorTranslator,
		v.translateError)
}

// NewValidator news a validator
func NewValidator() *Validator {
	return &Validator{
		valid: valid.New(),
		rules: make(map[string]*RegexValidateRule),
	}
}
