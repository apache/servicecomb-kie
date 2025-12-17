package kv

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-chassis/cari/db"
	"github.com/go-chassis/cari/db/config"
	_ "github.com/go-chassis/cari/db/etcd"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
)

func init() {
	cfg := config.Config{
		Kind:       "etcd",
		URI:        "http://127.0.0.1:2379",
		PoolSize:   0,
		TLSConfig:  nil,
		SSLEnabled: false,
		Timeout:    10 * time.Second,
	}
	cfg.Kind = "etcd"
	cfg.URI = "http://127.0.0.1:2379"
	cfg.Timeout = 10 * time.Second
	err := db.Init(&cfg)
	if err != nil {
		panic(err)
	}
}

func Test_listDataByNoCache(t *testing.T) {
	dao := Dao{}

	kvDoc1 := model.KVDoc{
		Key:     "Test_listDataByNoCache-application.yml",
		Value:   "test",
		Project: "default",
		Status:  "enabled",
		Labels: map[string]string{
			"app": "springCloud",
			"env": "dev",
		},
		Domain: "default",
	}
	kvDoc1.ID = fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join([]string{
		kvDoc1.Domain,
		kvDoc1.Project,
		kvDoc1.Key,
		kvDoc1.LabelFormat,
	}, "/"))))
	_, err := dao.Create(context.Background(), &kvDoc1)
	if err != nil {
		assert.True(t, errors.Is(datasource.ErrKVAlreadyExists, err))
	}

	kvDoc2 := model.KVDoc{
		Key:     "Test_listDataByNoCache-servicecomb.rateLimiting.test",
		Value:   "test",
		Project: "default",
		Status:  "enabled",
		Labels: map[string]string{
			"app": "springCloud",
			"env": "dev",
		},
		Domain: "default",
	}
	kvDoc2.ID = fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join([]string{
		kvDoc2.Domain,
		kvDoc2.Project,
		kvDoc2.Key,
		kvDoc2.LabelFormat,
	}, "/"))))
	_, err = dao.Create(context.Background(), &kvDoc2)
	if err != nil {
		assert.True(t, errors.Is(datasource.ErrKVAlreadyExists, err))
	}

	opts := datasource.FindOptions{
		Key: "beginWith(Test_listDataByNoCache-servicecomb.rateLimiting.)",
		Labels: map[string]string{
			"app": "springCloud",
			"env": "dev",
		},
	}
	re, reErr := toRegex(opts)
	assert.Nil(t, reErr)

	// onlyLabelFilteredResult 是仅通过label过滤后的值，没有后续的过滤条件，因此其内容应当更多
	result, onlyLabelFilteredResult, listErr := listDataByNoCache(context.Background(), "default", "default", re, opts)
	assert.Nil(t, listErr)
	assert.True(t, onlyLabelFilteredResult.Total >= 2)
	assert.True(t, result.Total >= 1)
	assert.True(t, onlyLabelFilteredResult.Total > result.Total)
}
