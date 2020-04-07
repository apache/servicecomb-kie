package validate_test

import (
	"testing"

	"github.com/apache/servicecomb-kie/pkg/validate"
	"github.com/stretchr/testify/assert"
)

type student struct {
	Name    string `validate:"kieTest"`
	Address string `validate:"alpha,min=2,max=4"`
}

func TestNewValidator(t *testing.T) {
	r := validate.NewRule("kieTest", `^[a-zA-Z0-9]*$`, nil)
	valid := validate.NewValidator()
	err := valid.RegisterRule(r)
	assert.Nil(t, err)
	assert.Nil(t, valid.AddErrorTranslation4Tag("min"))
	assert.Nil(t, valid.AddErrorTranslation4Tag("max"))

	s := &student{Name: "a1", Address: "abc"}
	err = valid.Validate(s)
	assert.Nil(t, err)

	s = &student{Name: "a1-", Address: "abc"}
	err = valid.Validate(s)
	assert.NotNil(t, err)
	t.Log(err)

	s = &student{Name: "a1", Address: "abcde"}
	err = valid.Validate(s)
	assert.NotNil(t, err)
	t.Log(err)
}
