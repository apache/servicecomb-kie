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

package track

import (
	"context"

	dmongo "github.com/go-chassis/cari/db/mongo"
	"github.com/go-chassis/openlog"
	"github.com/gofrs/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
	mmodel "github.com/apache/servicecomb-kie/server/datasource/mongo/model"
)

// Dao is the implementation
type Dao struct {
}

// CreateOrUpdate create a record or update exist record
// If revision and session_id exists then update else insert
func (s *Dao) CreateOrUpdate(ctx context.Context, detail *model.PollingDetail) (*model.PollingDetail, error) {
	collection := dmongo.GetClient().GetDB().Collection(mmodel.CollectionPollingDetail)
	queryFilter := bson.M{"revision": detail.Revision, "domain": detail.Domain, "session_id": detail.SessionID}
	res := collection.FindOne(ctx, queryFilter)
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			id, err := uuid.NewV4()
			if err != nil {
				return nil, err
			}
			detail.ID = id.String()
			_, err = collection.InsertOne(ctx, detail)
			if err != nil {
				return nil, err
			}
			return detail, nil
		}
		return nil, res.Err()
	}
	_, err := collection.UpdateOne(ctx, queryFilter, bson.D{{Key: "$set", Value: detail}})
	if err != nil {
		return nil, err
	}
	return detail, nil
}

// Get is to get a track data
func (s *Dao) GetPollingDetail(ctx context.Context, detail *model.PollingDetail) ([]*model.PollingDetail, error) {
	collection := dmongo.GetClient().GetDB().Collection(mmodel.CollectionPollingDetail)
	queryFilter := bson.M{"domain": detail.Domain}
	if detail.SessionID != "" {
		queryFilter["session_id"] = detail.SessionID
	}
	if detail.IP != "" {
		queryFilter["ip"] = detail.IP
	}
	if detail.UserAgent != "" {
		queryFilter["user_agent"] = detail.UserAgent
	}
	if detail.URLPath != "" {
		queryFilter["url_path"] = detail.URLPath
	}
	if detail.Revision != "" {
		queryFilter["revision"] = detail.Revision
	}
	cur, err := collection.Find(ctx, queryFilter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	if cur.Err() != nil {
		return nil, err
	}
	records := make([]*model.PollingDetail, 0)
	for cur.Next(ctx) {
		curRecord := &model.PollingDetail{}
		if err := cur.Decode(curRecord); err != nil {
			openlog.Error("decode to KVs error: " + err.Error())
			return nil, err
		}
		curRecord.Domain = ""
		records = append(records, curRecord)
	}
	if len(records) == 0 {
		return nil, datasource.ErrRecordNotExists
	}
	return records, nil
}
