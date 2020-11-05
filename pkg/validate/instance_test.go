package validate_test

import (
	"testing"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/validate"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	string32 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" //32
	string128 := string32 + string32 + string32 + string32
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
		Key:   string128 + "a",
		Value: "a",
	}
	assert.Error(t, validate.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a",
		Value: "",
	}
	assert.NoError(t, validate.Validate(kvDoc))

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
		Labels: map[string]string{string32 + "a": "a"},
	}
	assert.Error(t, validate.Validate(kvDoc))

	ListKVRe := &model.ListKVRequest{Project: "a", Domain: "a",
		Key:  "beginWith(a)",
	}
	assert.NoError(t, validate.Validate(ListKVRe))

	ListKVRe = &model.ListKVRequest{Project: "a", Domain: "a",
		Key:  "beginW(a)",
	}
	assert.Error(t, validate.Validate(ListKVRe))

	ListKVRe = &model.ListKVRequest{Project: "a", Domain: "a",
		Key:  "beginW()",
	}
	assert.Error(t, validate.Validate(ListKVRe))
}
