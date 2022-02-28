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

package counter

import (
	"context"
	"errors"

	"github.com/go-chassis/cari/db/mongo"
	"github.com/go-chassis/openlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-kie/server/datasource/mongo/model"
)

const revision = "revision_counter"

//Dao is the implementation
type Dao struct {
}

//GetRevision return current revision number
func (s *Dao) GetRevision(ctx context.Context, domain string) (int64, error) {
	collection := mongo.GetClient().GetDB().Collection(model.CollectionCounter)
	filter := bson.M{"name": revision, "domain": domain}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			msg := "operation timeout"
			openlog.Error(msg)
			return 0, errors.New(msg)
		}
		return 0, err
	}
	defer cur.Close(ctx)
	c := &Counter{}
	for cur.Next(ctx) {
		if err := cur.Decode(c); err != nil {
			openlog.Error("decode error: " + err.Error())
			return 0, err
		}
	}
	return c.Count, nil
}

//ApplyRevision increase revision number and return modified value
func (s *Dao) ApplyRevision(ctx context.Context, domain string) (int64, error) {
	collection := mongo.GetClient().GetDB().Collection(model.CollectionCounter)
	filter := bson.M{"name": revision, "domain": domain}
	sr := collection.FindOneAndUpdate(ctx, filter,
		bson.D{
			{Key: "$inc", Value: bson.D{
				{Key: "count", Value: 1},
			}}}, options.FindOneAndUpdate().SetReturnDocument(options.After))
	if sr.Err() != nil {
		return 0, sr.Err()
	}
	c := &Counter{}
	err := sr.Decode(c)
	if err != nil {
		openlog.Error("decode error: " + err.Error())
		return 0, err
	}
	return c.Count, nil
}
