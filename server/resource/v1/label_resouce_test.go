package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/apache/servicecomb-kie/pkg/model"
	handler2 "github.com/apache/servicecomb-kie/server/handler"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/go-chassis/go-chassis/core/common"
	"github.com/go-chassis/go-chassis/core/handler"
	"github.com/go-chassis/go-chassis/server/restful/restfultest"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLabelResource_PutLabel(t *testing.T) {
	t.Run("update label alias", func(t *testing.T) {
		kv := &model.KVDoc{
			Key:   "test",
			Value: "revisions",
			Labels: map[string]string{
				"test": "revisions",
			},
			Domain:  "default",
			Project: "test",
		}
		kv, _ = service.KVService.CreateOrUpdate(context.Background(), kv)
		j := []byte("{\"alias\":\"test\",\"id\":\"" + kv.LabelID + "\"}")
		r, _ := http.NewRequest("PUT", "/v1/test/kie/label", bytes.NewBuffer(j))
		r.Header.Add("Content-Type", "application/json")
		revision := &v1.LabelResource{}
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		c, err := restfultest.New(revision, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.LabelDoc{}
		err = json.Unmarshal(body, &data)
		assert.NoError(t, err)
		assert.Equal(t, data.Alias, "test")
	})
	t.Run("put label", func(t *testing.T) {
		label := &model.LabelDoc{
			Labels: map[string]string{
				"test": "revisions",
			},
			Domain:  "default",
			Project: "test",
		}
		j, _ := json.Marshal(label)
		r, _ := http.NewRequest("PUT", "/v1/test/kie/label", bytes.NewBuffer(j))
		r.Header.Add("Content-Type", "application/json")
		revision := &v1.LabelResource{}
		noopH := &handler2.NoopAuthHandler{}
		chain, _ := handler.CreateChain(common.Provider, "testchain1", noopH.Name())
		c, err := restfultest.New(revision, chain)
		assert.NoError(t, err)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		data := &model.LabelDoc{}
		err = json.Unmarshal(body, &data)
		assert.NoError(t, err)
		//assert.NotEmpty(t, data.ID)
	})
}
