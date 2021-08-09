package validator_test

import (
	"strings"
	"testing"

	"github.com/go-chassis/foundation/validator"

	"github.com/apache/servicecomb-kie/pkg/model"
	validsvc "github.com/apache/servicecomb-kie/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func init() {
	if err := validsvc.Init(); err != nil {
		panic(err)
	}
}

func TestValidate(t *testing.T) {
	string32 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" //32
	string128 := string32 + string32 + string32 + string32

	kvDoc := &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a",
		Value: "a",
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "",
		Value: "a",
	}
	assert.Error(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a#",
		Value: "a",
	}
	assert.Error(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   string128 + "a",
		Value: "a",
	}
	assert.Error(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a",
		Value: "",
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:       "a",
		Value:     "a",
		ValueType: "",
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:       "a",
		Value:     "a",
		ValueType: "text",
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:       "a",
		Value:     "a",
		ValueType: "a",
	}
	assert.Error(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Status: "",
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Status: "enabled",
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Status: "a",
	}
	assert.Error(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Labels: nil,
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Labels: map[string]string{"a": "a"},
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a",
		Value: "a",
		Labels: map[string]string{
			"1": "a",
			"2": "a",
			"3": "a",
			"4": "a",
			"5": "a",
			"6": "a",
			"7": "a", // error
		},
	}
	assert.Error(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a",
		Value: "a",
		Labels: map[string]string{
			"1":                            "a",
			"2":                            "a",
			"3":                            "a",
			"4":                            "a",
			"5":                            "a",
			"6-" + strings.Repeat("x", 31): "a", // error
		},
	}
	assert.Error(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a",
		Value: "a",
		Labels: map[string]string{
			"1": "a",
			"2": "a",
			"3": "a",
			"4": "a",
			"5": "a",
			"6": "a-" + strings.Repeat("x", 159), // error
		},
	}
	assert.Error(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Labels: map[string]string{string32 + "a": "a"},
	}
	assert.Error(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:       "a",
		ValueType: "123",
		Value:     "a",
		Labels:    map[string]string{"a": "a"},
	}
	err := validator.Validate(kvDoc)
	assert.Error(t, err, "validate failed, field: KVDoc.ValueType, rule: ^$|^(ini|json|text|yaml|properties|xml)$")

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:       "a",
		ValueType: "xml",
		Value:     "a",
		Labels:    map[string]string{"a": "a"},
	}
	assert.NoError(t, validator.Validate(kvDoc))

	ListKVRe := &model.ListKVRequest{Project: "a", Domain: "a",
		Key: "beginWith(a)",
	}
	assert.NoError(t, validator.Validate(ListKVRe))

	ListKVRe = &model.ListKVRequest{Project: "a", Domain: "a",
		Key: "beginW(a)",
	}
	assert.Error(t, validator.Validate(ListKVRe))

	ListKVRe = &model.ListKVRequest{Project: "a", Domain: "a",
		Key: "beginW()",
	}
	assert.Error(t, validator.Validate(ListKVRe))
}
