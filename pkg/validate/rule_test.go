package validate_test

import (
	"testing"

	"github.com/apache/servicecomb-kie/pkg/validate"
	"github.com/stretchr/testify/assert"
)

func TestNewRule(t *testing.T) {
	rule := validate.NewRule("t", `^[a-zA-Z0-9]*$`, &validate.Option{Min: 2, Max: 4})
	assert.Equal(t, "t", rule.Tag())
	rule.Explain()
	assert.True(t, rule.Validate("ab"))
	assert.False(t, rule.Validate("a"))
	assert.False(t, rule.Validate("abcde"))
	assert.False(t, rule.Validate("ab-"))

	rule = validate.NewRule("t", `^[a-zA-Z0-9]*$`, &validate.Option{Min: 2})
	rule.Explain()
	assert.True(t, rule.Validate("ab"))
	assert.False(t, rule.Validate("a"))
	assert.True(t, rule.Validate("abcde"))
	assert.False(t, rule.Validate("ab-"))

	rule = validate.NewRule("t", `^[a-zA-Z0-9]*$`, &validate.Option{Max: 4})
	rule.Explain()
	assert.True(t, rule.Validate("ab"))
	assert.True(t, rule.Validate("a"))
	assert.False(t, rule.Validate("abcde"))
	assert.False(t, rule.Validate("ab-"))

	rule = validate.NewRule("t", `^[a-zA-Z0-9]*$`, nil)
	rule.Explain()
	assert.True(t, rule.Validate("ab"))
	assert.True(t, rule.Validate("a"))
	assert.True(t, rule.Validate("abcdefg12345678"))
	assert.False(t, rule.Validate("ab-"))
}
