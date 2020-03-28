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
	Format  string            `bson:"format,omitempty"`
	Domain  string            `json:"domain,omitempty" yaml:"domain,omitempty"` //tenant info
	Project string            `json:"project,omitempty" yaml:"project,omitempty"`
	Alias   string            `json:"alias,omitempty" yaml:"alias,omitempty"`
}

//KVDoc is database struct to store kv
type KVDoc struct {
	ID             string `json:"id,omitempty" bson:"id,omitempty" yaml:"id,omitempty" swag:"string"`
	LabelID        string `json:"label_id,omitempty" bson:"label_id,omitempty" yaml:"label_id,omitempty"`
	Key            string `json:"key" yaml:"key" validate:"key"`
	Value          string `json:"value,omitempty" yaml:"value,omitempty" validate:"ascii,min=1,max=2097152"`
	ValueType      string `json:"value_type,omitempty" bson:"value_type,omitempty" yaml:"value_type,omitempty" validate:"valueType"` //ini,json,text,yaml,properties
	Checker        string `json:"check,omitempty" yaml:"check,omitempty"`                                                            //python script
	CreateRevision int64  `json:"create_revision,omitempty" bson:"create_revision," yaml:"create_revision,omitempty"`
	UpdateRevision int64  `json:"update_revision,omitempty" bson:"update_revision," yaml:"update_revision,omitempty"`
	Project        string `json:"project,omitempty" yaml:"project,omitempty" validate:"key"`
	Status         string `json:"status,omitempty" yaml:"status,omitempty" validate:"kvStatus"`
	CreateTime     int64  `json:"create_time,omitempty" bson:"create_time," yaml:"create_time,omitempty"`
	UpdateTime     int64  `json:"update_time,omitempty" bson:"update_time," yaml:"update_time,omitempty"`

	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" validate:"max=64,dive,keys,key,endkeys,key"` //redundant
	Domain string            `json:"domain,omitempty" yaml:"domain,omitempty" validate:"key"`                              //redundant

}

//ViewDoc is db struct, it saves user's custom view name and criteria
type ViewDoc struct {
	ID       string `json:"id,omitempty" bson:"id,omitempty" yaml:"id,omitempty" swag:"string"`
	Display  string `json:"display,omitempty" yaml:"display,omitempty"`
	Project  string `json:"project,omitempty" yaml:"project,omitempty"`
	Domain   string `json:"domain,omitempty" yaml:"domain,omitempty"`
	Criteria string `json:"criteria,omitempty" yaml:"criteria,omitempty"`
}

//PollingDetail is db struct, it record operation history
type PollingDetail struct {
	ID             string                 `json:"id,omitempty" yaml:"id,omitempty"`
	SessionID      string                 `json:"session_id,omitempty" bson:"session_id," yaml:"session_id,omitempty"`
	Domain         string                 `json:"domain,omitempty" yaml:"domain,omitempty"`
	PollingData    map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
	IP             string                 `json:"ip,omitempty" yaml:"ip,omitempty"`
	UserAgent      string                 `json:"user_agent,omitempty" bson:"user_agent," yaml:"user_agent,omitempty"`
	URLPath        string                 `json:"url_path,omitempty"  bson:"url_path,"  yaml:"url_path,omitempty"`
	ResponseBody   interface{}            `json:"response_body,omitempty"  bson:"response_body,"  yaml:"response_body,omitempty"`
	ResponseHeader map[string][]string    `json:"response_header,omitempty"  bson:"response_header,"  yaml:"response_header,omitempty"`
	ResponseCode   int                    `json:"response_code,omitempty"  bson:"response_code,"  yaml:"response_code,omitempty"`
}
