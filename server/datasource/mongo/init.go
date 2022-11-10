/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongo

import (
	"context"

	"github.com/apache/servicecomb-kie/server/datasource/mongo/rbac"
	rbacdao "github.com/apache/servicecomb-kie/server/datasource/rbac"
	dmongo "github.com/go-chassis/cari/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/counter"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/history"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/kv"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/model"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/track"
)

type Broker struct {
}

func NewFrom(c *datasource.Config) (datasource.Broker, error) {
	broker := Broker{}
	err := ensureDB()
	if err != nil {
		return nil, err
	}
	return &broker, nil
}
func (*Broker) GetRevisionDao() datasource.RevisionDao {
	return &counter.Dao{}
}
func (*Broker) GetKVDao() datasource.KVDao {
	return &kv.Dao{}
}
func (*Broker) GetHistoryDao() datasource.HistoryDao {
	return &history.Dao{}
}
func (*Broker) GetTrackDao() datasource.TrackDao {
	return &track.Dao{}
}
func (*Broker) GetRbacDao() rbacdao.Dao {
	return &rbac.Dao{}
}

func ensureDB() error {
	err := ensureRevisionCounter()
	ensureKV()
	ensureKVRevision()
	ensureView()
	ensureKVLongPolling()
	return err
}

func ensureRevisionCounter() error {
	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"name", "domain", "count"},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	revisionCounterIndex := buildIndexDoc("name", "domain")
	revisionCounterIndex.Options = options.Index().SetUnique(true)
	dmongo.EnsureCollection(model.CollectionCounter, validator, []mongo.IndexModel{revisionCounterIndex})
	_, err := dmongo.GetClient().GetDB().Collection(model.CollectionCounter).UpdateOne(context.Background(),
		bson.M{"name": "revision_counter", "domain": "default"},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "count", Value: 1},
			}},
		}, options.Update().SetUpsert(true))
	return err
}

func ensureKV() {
	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"key", "domain", "project", "id"},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	kvIndex := buildIndexDoc("id")
	kvIndex.Options = options.Index().SetUnique(true)
	dmongo.EnsureCollection(model.CollectionKV, validator, []mongo.IndexModel{kvIndex})
}

func ensureKVRevision() {
	kvRevisionIndex := buildIndexDoc("delete_time")
	kvRevisionIndex.Options = options.Index().SetExpireAfterSeconds(7 * 24 * 3600)
	dmongo.EnsureCollection(model.CollectionKVRevision, nil, []mongo.IndexModel{kvRevisionIndex})
}

func ensureView() {
	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"id", "domain", "project", "display"},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	viewIDIndex := buildIndexDoc("id")
	viewIDIndex.Options = options.Index().SetUnique(true)
	viewMultipleIndex := buildIndexDoc("display", "domain", "project")
	viewMultipleIndex.Options = options.Index().SetUnique(true)
	dmongo.EnsureCollection(model.CollectionView, validator, []mongo.IndexModel{viewIDIndex, viewMultipleIndex})
}

func ensureKVLongPolling() {
	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"id", "revision", "session_id", "url_path"},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	timestampIndex := buildIndexDoc("timestamp")
	timestampIndex.Options = options.Index().SetExpireAfterSeconds(7 * 24 * 3600)
	kvLongPollingIndex := buildIndexDoc("revision", "domain", "session_id")
	kvLongPollingIndex.Options = options.Index().SetUnique(true)
	dmongo.EnsureCollection(model.CollectionPollingDetail, validator, []mongo.IndexModel{timestampIndex, kvLongPollingIndex})
}

func buildIndexDoc(keys ...string) mongo.IndexModel {
	keysDoc := bsonx.Doc{}
	for _, key := range keys {
		keysDoc = keysDoc.Append(key, bsonx.Int32(1))
	}
	index := mongo.IndexModel{
		Keys: keysDoc,
	}
	return index
}
func init() {
	datasource.RegisterPlugin("mongo", NewFrom)
}
