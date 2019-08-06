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
	"fmt"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/db"
	"github.com/apache/servicecomb-kie/server/service/label"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson"
)

//GetAndAddHistory get latest labels revision and call AddHistory
func GetAndAddHistory(ctx context.Context,
	labelID string, labels map[string]string, kvs []*model.KVDoc, domain string, project string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, db.Timeout)
	defer cancel()
	r, err := label.GetLatestLabel(ctx, labelID)
	if err != nil {
		if err == db.ErrRevisionNotExist {
			openlogging.Warn(fmt.Sprintf("label revision not exists, create first label revision"))
			r = &model.LabelRevisionDoc{
				LabelID:  labelID,
				Labels:   labels,
				Domain:   domain,
				Revision: 0,
			}
		} else {
			openlogging.Error(fmt.Sprintf("get latest [%s] in [%s],err: %s",
				labelID, domain, err.Error()))
			return 0, err
		}

	}
	r.Revision, err = AddHistory(ctx, r, labelID, kvs)
	if err != nil {
		return 0, err
	}
	return r.Revision, nil
}

//GetHistoryByLabelID get all history by label id
func GetHistoryByLabelID(ctx context.Context, labelID string) ([]*model.LabelRevisionDoc, error) {
	ctx, cancel := context.WithTimeout(ctx, db.Timeout)
	defer cancel()
	filter := bson.M{"label_id": labelID}
	return getHistoryByLabelID(ctx, filter)
}
