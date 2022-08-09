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

import "time"

// LabelDoc is database struct to store labels
type LabelDoc struct {
	ID      string            `json:"id,omitempty" bson:"id,omitempty" yaml:"id,omitempty" swag:"string"`
	Labels  map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Format  string            `bson:"format,omitempty"`
	Domain  string            `json:"domain,omitempty" yaml:"domain,omitempty"` //tenant info
	Project string            `json:"project,omitempty" yaml:"project,omitempty"`
	Alias   string            `json:"alias,omitempty" yaml:"alias,omitempty"`
}

// KVDoc is database struct to store kv
type KVDoc struct {
	ID             string `json:"id,omitempty" bson:"id,omitempty" yaml:"id,omitempty" swag:"string"`
	LabelFormat    string `json:"label_format,omitempty" bson:"label_format,omitempty" yaml:"label_format,omitempty"`
	Key            string `json:"key" yaml:"key" validate:"min=1,max=2048,key"`
	Value          string `json:"value" yaml:"value" validate:"max=131072,value"`                                                    //128K
	ValueType      string `json:"value_type,omitempty" bson:"value_type,omitempty" yaml:"value_type,omitempty" validate:"valueType"` //ini,json,text,yaml,properties,xml
	Checker        string `json:"check,omitempty" yaml:"check,omitempty" validate:"max=1048576,check"`                               //python script
	CreateRevision int64  `json:"create_revision,omitempty" bson:"create_revision," yaml:"create_revision,omitempty"`
	UpdateRevision int64  `json:"update_revision,omitempty" bson:"update_revision," yaml:"update_revision,omitempty"`
	Project        string `json:"project,omitempty" yaml:"project,omitempty" validate:"min=1,max=256,commonName"`
	Status         string `json:"status,omitempty" yaml:"status,omitempty" validate:"kvStatus"`
	CreateTime     int64  `json:"create_time,omitempty" bson:"create_time," yaml:"create_time,omitempty"`
	UpdateTime     int64  `json:"update_time,omitempty" bson:"update_time," yaml:"update_time,omitempty"`

	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" validate:"max=6,dive,keys,labelK,endkeys,labelV"` //redundant
	Domain string            `json:"domain,omitempty" yaml:"domain,omitempty" validate:"min=1,max=256,commonName"`              //redundant
}

// ViewDoc is db struct, it saves user's custom view name and criteria
type ViewDoc struct {
	ID       string `json:"id,omitempty" bson:"id,omitempty" yaml:"id,omitempty" swag:"string"`
	Display  string `json:"display,omitempty" yaml:"display,omitempty"`
	Project  string `json:"project,omitempty" yaml:"project,omitempty"`
	Domain   string `json:"domain,omitempty" yaml:"domain,omitempty"`
	Criteria string `json:"criteria,omitempty" yaml:"criteria,omitempty"`
}

// PollingDetail is db struct, it record operation history
type PollingDetail struct {
	ID           string                 `json:"id,omitempty" yaml:"id,omitempty"`
	SessionID    string                 `json:"session_id,omitempty" bson:"session_id," yaml:"session_id,omitempty"`
	SessionGroup string                 `json:"session_group,omitempty" bson:"session_group," yaml:"session_group,omitempty"`
	Domain       string                 `json:"domain,omitempty" yaml:"domain,omitempty"`
	Project      string                 `json:"project,omitempty" yaml:"project,omitempty"`
	PollingData  map[string]interface{} `json:"polling_data,omitempty" yaml:"polling_data,omitempty"`
	Revision     string                 `json:"revision,omitempty" yaml:"revision,omitempty"`
	IP           string                 `json:"ip,omitempty" yaml:"ip,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty" bson:"user_agent," yaml:"user_agent,omitempty"`
	URLPath      string                 `json:"url_path,omitempty"  bson:"url_path,"  yaml:"url_path,omitempty"`
	ResponseBody []*KVDoc               `json:"kv,omitempty"  bson:"kv,"  yaml:"kv,omitempty"`
	ResponseCode int                    `json:"response_code,omitempty"  bson:"response_code,"  yaml:"response_code,omitempty"`
	Timestamp    time.Time              `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
}

// UpdateKVRequest is db struct, it contains kv update request params
type UpdateKVRequest struct {
	ID      string `json:"id,omitempty" bson:"id,omitempty" yaml:"id,omitempty" swag:"string" validate:"uuid"`
	Value   string `json:"value,omitempty" yaml:"value,omitempty" validate:"max=131072,value"`
	Project string `json:"project,omitempty" yaml:"project,omitempty" validate:"min=1,max=256,commonName"`
	Domain  string `json:"domain,omitempty" yaml:"domain,omitempty" validate:"min=1,max=256,commonName"` //redundant
	Status  string `json:"status,omitempty" yaml:"status,omitempty" validate:"kvStatus"`
}

// GetKVRequest contains kv get request params
type GetKVRequest struct {
	Project string `json:"project,omitempty" yaml:"project,omitempty" validate:"min=1,max=256,commonName"`
	Domain  string `json:"domain,omitempty" yaml:"domain,omitempty" validate:"min=1,max=256,commonName"` //redundant
	ID      string `json:"id,omitempty" bson:"id,omitempty" yaml:"id,omitempty" swag:"string" validate:"uuid"`
}

// ListKVRequest contains kv list request params
type ListKVRequest struct {
	Project string            `json:"project,omitempty" yaml:"project,omitempty" validate:"min=1,max=256,commonName"`
	Domain  string            `json:"domain,omitempty" yaml:"domain,omitempty" validate:"min=1,max=256,commonName"` //redundant
	Key     string            `json:"key" yaml:"key" validate:"max=128,getKey"`
	Labels  map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" validate:"max=8,dive,keys,labelK,endkeys,labelV"` //redundant
	Offset  int64             `validate:"min=0"`
	Limit   int64             `validate:"min=0,max=100"`
	Status  string            `json:"status,omitempty" yaml:"status,omitempty" validate:"kvStatus"`
	Match   string            `json:"match,omitempty" yaml:"match,omitempty"`
}

// UploadKVRequest contains kv list upload request params
type UploadKVRequest struct {
	Domain   string `json:"domain,omitempty" yaml:"domain,omitempty" validate:"min=1,max=256,commonName"` //redundant
	Project  string `json:"project,omitempty" yaml:"project,omitempty" validate:"min=1,max=256,commonName"`
	KVs      []*KVDoc
	Override string
}
