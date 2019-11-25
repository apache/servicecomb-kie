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

package kv

import (
	"context"
	"github.com/apache/servicecomb-kie/server/service"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/mongo"
)

//clearAll clean attr which don't need to return to client side
func clearAll(kv *model.KVDoc) {
	clearPart(kv)
	kv.Labels = nil
	kv.LabelID = ""
}
func clearPart(kv *model.KVDoc) {
	kv.Domain = ""
	kv.Project = ""
}
func cursorToOneKV(ctx context.Context, cur *mongo.Cursor, labels map[string]string) ([]*model.KVResponse, error) {
	result := make([]*model.KVResponse, 0)
	curKV := &model.KVDoc{} //reuse this pointer to reduce GC, only clear label
	//check label length to get the exact match
	for cur.Next(ctx) { //although complexity is O(n), but there won't be so much labels for one key
		if cur.Err() != nil {
			return nil, cur.Err()
		}
		curKV.Labels = nil
		err := cur.Decode(curKV)
		if err != nil {
			openlogging.Error("decode error: " + err.Error())
			return nil, err
		}
		if len(curKV.Labels) == len(labels) {
			openlogging.Debug("hit exact labels")
			labelGroup := &model.KVResponse{
				LabelDoc: &model.LabelDocResponse{
					Labels:  labels,
					LabelID: curKV.LabelID,
				},
				Data: make([]*model.KVDoc, 0),
			}
			clearAll(curKV)
			labelGroup.Data = append(labelGroup.Data, curKV)
			result = append(result, labelGroup)
			return result, nil
		}

	}
	return nil, service.ErrKeyNotExists
}
