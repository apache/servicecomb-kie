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

package history_test

import (
	"context"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func init() {
	config.Configurations = &config.Config{DB: config.DB{URI: "mongodb://kie:123@127.0.0.1:27017/kie"}}
	_ = session.Init()
}

func TestAddHistory(t *testing.T) {
	ctx := context.Background()
	coll := session.GetDB().Collection("label_revision")
	cur, err := coll.Find(
		context.Background(),
		bson.M{
			"label_format": "5dbc079183ff1a09242376e7",
			"data.key":     "lb",
		})
	assert.NoError(t, err)
	for cur.Next(ctx) {
		var elem interface{}
		err := cur.Decode(&elem)
		assert.NoError(t, err)
		t.Log(elem)
	}
}
