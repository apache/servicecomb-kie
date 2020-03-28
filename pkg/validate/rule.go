package validate

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/go-playground/validator"
)

// RegexValidateRule contains an validate tag's info
type RegexValidateRule struct {
	tag           string
	min           int64
	max           int64
	regex         *regexp.Regexp
	validateFuncs []func(string) bool
}

// Option is RegexValidateRule option
type Option struct {
	Min int64
	Max int64
}

// Validate validates string
func (r *RegexValidateRule) Validate(s string) bool {
	for _, f := range r.validateFuncs {
		if ok := f(s); !ok {
			return false
		}
	}
	return true
}

func (r *RegexValidateRule) validateFL(fl validator.FieldLevel) bool {
	return r.Validate(fl.Field().String())
}

// Tag returns the validate rule's tag
func (r *RegexValidateRule) Tag() string {
	return r.tag
}

// Explain explains the rule
func (r *RegexValidateRule) Explain() string {
	explain := r.regex.String()
	if r.max > 0 {
		explain = fmt.Sprintf("%s , max = %d", explain, r.max)
	}
	if r.min > 0 {
		explain = fmt.Sprintf("%s , min = %d", explain, r.min)
	}
	return explain
}

func (r *RegexValidateRule) matchRegex(s string) bool {
	return r.regex.MatchString(s)
}
func (r *RegexValidateRule) matchMin(s string) bool {
	return int64(utf8.RuneCountInString(s)) >= r.min
}
func (r *RegexValidateRule) matchMax(s string) bool {
	return int64(utf8.RuneCountInString(s)) <= r.max
}

// NewRule news a rule
func NewRule(tag, regexStr string, opt *Option) *RegexValidateRule {
	r := &RegexValidateRule{
		tag:           tag,
		regex:         regexp.MustCompile(regexStr),
		validateFuncs: make([]func(string) bool, 0),
	}

	if opt == nil {
		r.validateFuncs = append(r.validateFuncs, r.matchRegex)
		return r
	}

	if opt.Max > 0 {
		r.max = opt.Max
		r.validateFuncs = append(r.validateFuncs, r.matchMax)
	}
	if opt.Min > 0 {
		r.min = opt.Min
		r.validateFuncs = append(r.validateFuncs, r.matchMin)
	}
	r.validateFuncs = append(r.validateFuncs, r.matchRegex)
	return r
}
