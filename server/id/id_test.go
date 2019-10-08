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

package id_test

import (
	"github.com/apache/servicecomb-kie/server/id"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestID_MarshalBSONValue(t *testing.T) {
	type Obj struct {
		ID id.ID `bson:"label_id,omitempty"`
	}

	o := new(Obj)
	o.ID = id.ID(primitive.NewObjectID().Hex())

	b, err := bson.Marshal(o)
	assert.NoError(t, err)
	t.Log(b)

	o2 := new(Obj)
	err = bson.Unmarshal(b, o2)
	assert.NoError(t, err)
	t.Log(o2)
}
