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

//KVRequest is http request body
type KVRequest struct {
	Key       string            `json:"key" yaml:"key"`
	Value     string            `json:"value,omitempty" yaml:"value,omitempty"`
	ValueType string            `json:"value_type,omitempty" bson:"value_type,omitempty" yaml:"value_type,omitempty"` //ini,json,text,yaml,properties
	Checker   string            `json:"check,omitempty" yaml:"check,omitempty"`                                       //python script
	Labels    map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`                                     //redundant
}

//KVResponse represents the key value list
type KVResponse struct {
	LabelDoc *LabelDocResponse `json:"label,omitempty"`
	Total    int               `json:"total,omitempty"`
	Data     []*KVDoc          `json:"data,omitempty"`
}

//LabelDocResponse is label struct
type LabelDocResponse struct {
	LabelID string            `json:"label_id,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}

//LabelHistoryResponse is label history revision struct
type LabelHistoryResponse struct {
	LabelID  string            `json:"label_id,omitempty"  bson:"label_id,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
	KVs      []*KVDoc          `json:"data,omitempty"`
	Revision int               `json:"revision"`
}

//ViewResponse represents the view list
type ViewResponse struct {
	Total int        `json:"total,omitempty"`
	Data  []*ViewDoc `json:"data,omitempty"`
}

//DocResponseSingleKey is response doc
type DocResponseSingleKey struct {
	CreateRevision int64             `json:"create_revision"`
	CreateTime     string            `json:"create_time"`
	ID             string            `json:"id"`
	Key            string            `json:"key"`
	LabelID        string            `json:"label_id"`
	Labels         map[string]string `json:"labels"`
	UpdateRevision int64             `json:"update_revision"`
	UpdateTime     string            `json:"update_time"`
	Value          string            `json:"value"`
	ValueType      string            `json:"value_type"`
}

//DocResponseGetKey is response doc
type DocResponseGetKey struct {
	Data  []*DocResponseSingleKey `json:"data"`
	Total int64                   `json:"total"`
}

//PollingDataResponse  is response doc
type PollingDataResponse struct {
	Data  []*PollingDetail `json:"data"`
	Total int              `json:"total"`
}

//DocHealthCheck is response doc
type DocHealthCheck struct {
	Version   string `json:"version"`
	Revision  string `json:"revision"`
	Timestamp int64  `json:"timestamp"`
}
