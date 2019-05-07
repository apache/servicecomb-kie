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

package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Labels map[string]string

type KV struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Key       string             `json:"key"`
	Value     string             `json:"value"`
	ValueType string             `json:"valueType"`        //ini,json,text,yaml,properties
	Domain    string             `json:"domain"`           //tenant info
	Labels    map[string]string  `json:"labels,omitempty"` //key has labels
	Checker   string             `json:"check,omitempty"`  //python script
	Revision  int                `json:"revision"`
}
type KVHistory struct {
	KID      string `json:"id,omitempty" bson:"kvID"`
	Value    string `json:"value"`
	Checker  string `json:"check,omitempty"` //python script
	Revision int    `json:"revision"`
}
