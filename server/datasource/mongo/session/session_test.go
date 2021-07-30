package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/session"
	_ "github.com/apache/servicecomb-kie/test"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestGetColInfo(t *testing.T) {
	var err error
	err = session.Init(&datasource.Config{
		URI:     "mongodb://kie:123@127.0.0.1:27017/kie",
		Timeout: 10 * time.Second,
	})
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
