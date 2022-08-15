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

var string32 = strings.Repeat("a", 32)
var string2048 = strings.Repeat("a", 2048)
var string131072 = strings.Repeat("a", 131072)

func TestValidate(t *testing.T) {
	kvDoc := &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a",
		Value: "a",
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a",
		Value: "",
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a",
		Value: string131072,
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "a",
		Value: string131072 + "a",
	}
	assert.Error(t, validator.Validate(kvDoc))
}

func TestKey(t *testing.T) {
	kvDoc := &model.KVDoc{Project: "a", Domain: "a",
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
		Key:   string2048 + "a",
		Value: "a",
	}
	assert.Error(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "zZ12.-_:",
		Value: "zZ12.-_:",
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "...zZ12.-_:",
		Value: "......asdfakdjlkaj;eje#$@%$RE$5zZ12.-_:",
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "_...zZ12.-_:",
		Value: "adslfjkla",
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:   "-_...zZ12.-_:",
		Value: "adslfjkla",
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

func TestLabels(t *testing.T) {
	kvDoc := &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Labels: nil,
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Labels: map[string]string{"": ""},
	}
	assert.Error(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Labels: map[string]string{"a": "a"},
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Labels: map[string]string{"a": ""},
	}
	assert.NoError(t, validator.Validate(kvDoc))

	kvDoc = &model.KVDoc{Project: "a", Domain: "a",
		Key:    "a",
		Value:  "a",
		Labels: map[string]string{"": "a"},
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
}

func TestValueType(t *testing.T) {
	kvDoc := &model.KVDoc{Project: "a", Domain: "a",
		Key:       "a",
		Value:     "a",
		ValueType: "text",
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
		ValueType: "a",
	}
	assert.Error(t, validator.Validate(kvDoc))
}

func TestStatus(t *testing.T) {
	kvDoc := &model.KVDoc{Project: "a", Domain: "a",
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
}

func TestGetKey(t *testing.T) {
	listKvreq := model.ListKVRequest{
		Project: "default",
		Domain:  "default",
		Key:     "beginWith(zZ12.-_:)",
		Labels: map[string]string{
			"service": "utService",
		},
		Offset: 0,
		Limit:  10,
		Status: "enabled",
		Match:  "exact",
	}
	assert.NoError(t, validator.Validate(listKvreq))

	listKvreq = model.ListKVRequest{
		Project: "default",
		Domain:  "default",
		Key:     "wildcard(*IME*)",
		Labels: map[string]string{
			"service": "utService",
		},
		Offset: 0,
		Limit:  10,
		Status: "enabled",
		Match:  "exact",
	}
	assert.NoError(t, validator.Validate(listKvreq))

	listKvreq = model.ListKVRequest{
		Project: "default",
		Domain:  "default",
		Key:     "zZ12.-_:",
		Labels: map[string]string{
			"service": "utService",
		},
		Offset: 0,
		Limit:  10,
		Status: "enabled",
		Match:  "exact",
	}
	assert.NoError(t, validator.Validate(listKvreq))

	listKvreq = model.ListKVRequest{
		Project: "default",
		Domain:  "default",
		Key:     "wildcard(*zZ12.-_:*)",
		Labels: map[string]string{
			"service": "utService",
		},
		Offset: 0,
		Limit:  10,
		Status: "enabled",
		Match:  "exact",
	}
	assert.NoError(t, validator.Validate(listKvreq))
}
