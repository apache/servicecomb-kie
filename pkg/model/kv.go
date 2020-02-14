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

//PollingDetail record operation history
type PollingDetail struct {
	ID             string                 `json:"id,omitempty" yaml:"id,omitempty"`
	SessionID      string                 `json:"session_id,omitempty" yaml:"session_id,omitempty"`
	Domain         string                 `json:"domain,omitempty" yaml:"domain,omitempty"`
	PollingData    map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
	IP             string                 `json:"ip,omitempty" yaml:"ip,omitempty"`
	UserAgent      string                 `json:"user_agent,omitempty" yaml:"user_agent,omitempty"`
	URLPath        string                 `json:"url_path,omitempty" yaml:"url_path,omitempty"`
	ResponseBody   interface{}            `json:"response_body,omitempty" yaml:"response_body,omitempty"`
	ResponseHeader map[string][]string    `json:"response_header,omitempty" yaml:"response_header,omitempty"`
	ResponseCode   int                    `json:"response_code,omitempty" yaml:"response_code,omitempty"`
}
