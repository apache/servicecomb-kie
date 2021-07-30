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

package view

import (
	"context"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/session"
	"github.com/go-chassis/openlog"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func create(ctx context.Context, viewDoc *model.ViewDoc) error {
	viewDoc.ID = uuid.NewV4().String()
	viewDoc.Criteria = "" //TODO parse pipe line to sql-like lang
	_, err := session.GetDB().Collection(session.CollectionView).InsertOne(ctx, viewDoc)
	if err != nil {
		openlog.Error("can not insert view collection: " + err.Error())
		return session.ErrViewCreation
	}
	return nil
}
func findOne(ctx context.Context, viewID, domain, project string) (*model.ViewDoc, error) {
	filter := bson.M{"domain": domain,
		"project": project,
		"id":      viewID}
	sr := session.GetDB().Collection(session.CollectionView).FindOne(ctx, filter)
	if sr.Err() != nil {
		openlog.Error("can not find view collection: " + sr.Err().Error())
		return nil, sr.Err()
	}
	result := &model.ViewDoc{}
	err := sr.Decode(result)
	if err != nil {
		openlog.Error("decode error: " + err.Error())
		return nil, err
	}
	if result.ID == viewID {
		return result, nil
	}
	return nil, session.ErrViewNotExist
}
