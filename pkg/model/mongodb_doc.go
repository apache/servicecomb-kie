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

//LabelDoc is database struct to store labels
type LabelDoc struct {
	ID      string            `json:"id,omitempty" bson:"id,omitempty" yaml:"id,omitempty" swag:"string"`
	Labels  map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Domain  string            `json:"domain,omitempty" yaml:"domain,omitempty"` //tenant info
	Project string            `json:"project,omitempty" yaml:"project,omitempty"`
}

//KVDoc is database struct to store kv
type KVDoc struct {
	ID        string `json:"id,omitempty" bson:"id,omitempty" yaml:"id,omitempty" swag:"string"`
	LabelID   string `json:"label_id,omitempty" bson:"label_id,omitempty" yaml:"label_id,omitempty"`
	Key       string `json:"key" yaml:"key"`
	Value     string `json:"value,omitempty" yaml:"value,omitempty"`
	ValueType string `json:"value_type,omitempty" bson:"value_type,omitempty" yaml:"value_type,omitempty"` //ini,json,text,yaml,properties
	Checker   string `json:"check,omitempty" yaml:"check,omitempty"`                                       //python script

	Labels   map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"` //redundant
	Domain   string            `json:"domain,omitempty" yaml:"domain,omitempty"` //redundant
	Revision int               `json:"revision,omitempty" bson:"revision," yaml:"revision,omitempty"`
	Project  string            `json:"project,omitempty" yaml:"project,omitempty"`
}
