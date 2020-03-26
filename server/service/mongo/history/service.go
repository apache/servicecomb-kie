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

package history

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Service is the implementation
type Service struct {
}

//GetHistory get all history by label id
func (s *Service) GetHistory(ctx context.Context, kvID string, options ...service.FindOption) (*model.KVResponse, error) {
	var filter primitive.M
	opts := service.FindOptions{}
	for _, o := range options {
		o(&opts)
	}
	filter = bson.M{
		"id": kvID,
	}

	return getHistoryByKeyID(ctx, filter, opts.Offset, opts.Limit)
}
