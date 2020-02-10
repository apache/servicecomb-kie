package session_test

import (
	"context"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestGetColInfo(t *testing.T) {
	var err error
	config.Configurations = &config.Config{DB: config.DB{URI: "mongodb://kie:123@127.0.0.1:27017/kie"}}
	err = session.Init()
	assert.NoError(t, err)
	err = session.CreateView(context.Background(), "test_view", session.CollectionKV, []bson.D{
		{{
			"$match",
			bson.D{{"domain", "default"}, {"project", "default"}},
		}},
	})
	assert.NoError(t, err)
	c, err := session.GetColInfo(context.Background(), "test_view")
	assert.NoError(t, err)
	assert.Equal(t, "default", c.Options.Pipeline[0]["$match"]["domain"])
	err = session.DropView(context.Background(), "test_view")
	assert.NoError(t, err)
}
