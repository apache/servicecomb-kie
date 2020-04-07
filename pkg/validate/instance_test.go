package validate_test

import (
	"testing"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/validate"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	s := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" //64
	testStr := s + s + s + s                                                //256
	err := validate.Init()
	assert.NoError(t, err)

	kvDoc := &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a",
		Value: "a",
	}
	assert.NoError(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "",
		Value: "a",
	}
	assert.Error(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a#",
		Value: "a",
	}
	assert.Error(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   testStr + "a",
		Value: "a",
	}
	assert.Error(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a",
		Value: "",
	}
	assert.Error(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:       "a",
		Value:     "a",
		ValueType: "",
	}
	assert.NoError(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:       "a",
		Value:     "a",
		ValueType: "text",
	}
	assert.NoError(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:       "a",
		Value:     "a",
		ValueType: "a",
	}
	assert.Error(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Status: "",
	}
	assert.NoError(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Status: "enabled",
	}
	assert.NoError(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Status: "a",
	}
	assert.Error(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Labels: nil,
	}
	assert.NoError(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Labels: map[string]string{"a": "a"},
	}
	assert.NoError(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Labels: map[string]string{testStr + "a": "a"},
	}
	assert.Error(t, validate.Validate(kvDoc))
}
