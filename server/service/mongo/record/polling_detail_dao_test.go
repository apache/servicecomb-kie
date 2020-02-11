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

package record_test

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/service/mongo/record"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateRecord(t *testing.T) {
	var err error
	config.Configurations = &config.Config{DB: config.DB{URI: "mongodb://kie:123@127.0.0.1:27017/kie"}}
	err = session.Init()
	assert.NoError(t, err)
	d, _ := record.CreateRecord(context.TODO(), &model.PollingDetail{
		ID:        uuid.NewV4().String(),
		IP:        "0.0.0.0",
		UserAgent: "xxx",
		URLPath:   "xx/xx",
	})
	//assert.NoError(t, err)
	assert.NotEmpty(t, d.ID)
}
