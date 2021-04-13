package cipherutil_test

import (
	"testing"

	"github.com/apache/servicecomb-kie/pkg/cipherutil"
	_ "github.com/apache/servicecomb-kie/test"
	"github.com/stretchr/testify/assert"
)

func TestTryDecrypt(t *testing.T) {
	t.Run("try decrypt failed, should return src", func(t *testing.T) {
		assert.Equal(t, "abc", cipherutil.TryDecrypt("abc"))
	})
}
