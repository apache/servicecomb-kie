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

package id

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

//ID decouple mongodb
type ID string

//UnmarshalBSONValue is implement
func (id *ID) UnmarshalBSONValue(t bsontype.Type, raw []byte) error {
	if t == bsontype.ObjectID && len(raw) == 12 {
		var objID primitive.ObjectID
		copy(objID[:], raw)
		*id = ID(objID.Hex())
		return nil
	} else if t == bsontype.String {
		if str, _, ok := bsoncore.ReadString(raw); ok {
			*id = ID(str)
			return nil
		}
	}

	return fmt.Errorf("unable to unmarshal bson id &mdash; type: %v, length: %v", len(raw), t)
}

//String return string
func (id ID) String() string {
	return string(id)
}
