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

package dao

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *MongodbService) createLabel(ctx context.Context, domain string, labels map[string]string) (*model.LabelDoc, error) {
	l := &model.LabelDoc{
		Domain: domain,
		Labels: labels,
	}
	collection := s.c.Database(DB).Collection(CollectionLabel)
	res, err := collection.InsertOne(ctx, l)
	if err != nil {
		return nil, err
	}
	objectID, _ := res.InsertedID.(primitive.ObjectID)
	l.ID = objectID
	return l, nil
}
func (s *MongodbService) findOneLabels(ctx context.Context, filter bson.M) (*model.LabelDoc, error) {
	collection := s.c.Database(DB).Collection(CollectionLabel)
	ctx, _ = context.WithTimeout(context.Background(), s.timeout)
	sr := collection.FindOne(ctx, filter)
	if sr.Err() != nil {
		return nil, sr.Err()
	}
	l := &model.LabelDoc{}
	err := sr.Decode(l)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrLabelNotExists
		}
		openlogging.Error("decode error: " + err.Error())
		return nil, err
	}
	return l, nil
}

//LabelsExist check label exists or not and return label ID
func (s *MongodbService) LabelsExist(ctx context.Context, domain string, labels map[string]string) (primitive.ObjectID, error) {
	l, err := s.FindLabels(ctx, domain, labels)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return l.ID, nil

}
