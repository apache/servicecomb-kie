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

package pubsub

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/stringutil"
	"github.com/apache/servicecomb-kie/pkg/util"
)

// const
const (
	ActionPut    = "put"
	ActionDelete = "del"
)

//KVChangeEvent is event between kie nodes, and broadcast by serf
type KVChangeEvent struct {
	Key      string
	Action   string //include: put,delete
	Labels   map[string]string
	DomainID string
	Project  string
}

func (e *KVChangeEvent) String() string {
	return strings.Join([]string{e.Key, e.Action, stringutil.FormatMap(e.Labels), e.DomainID, e.Project}, ";;")
}

//NewKVChangeEvent create a struct base on event payload
func NewKVChangeEvent(payload []byte) (*KVChangeEvent, error) {
	ke := &KVChangeEvent{}
	err := json.Unmarshal(payload, ke)
	return ke, err
}

//Topic can be subscribe
type Topic struct {
	Labels       map[string]string `json:"-"`
	LabelsFormat string            `json:"labels,omitempty"`
	DomainID     string            `json:"domainID,omitempty"`
	Project      string            `json:"project,omitempty"`
	MatchType    string            `json:"match,omitempty"`
}

func (t *Topic) Encode() (string, error) {
	t.LabelsFormat = stringutil.FormatMap(t.Labels)
	b, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

//ParseTopic parse topic string to topic struct
func ParseTopic(s string) (*Topic, error) {
	t := &Topic{
		Labels: make(map[string]string),
	}
	err := json.Unmarshal([]byte(s), t)
	if err != nil {
		return nil, err
	}
	if t.LabelsFormat == stringutil.LabelNone {
		return t, nil
	}
	ls := strings.Split(t.LabelsFormat, "::")
	if len(ls) != 0 {
		for _, l := range ls {
			s := strings.Split(l, "=")
			if len(s) != 2 {
				return nil, errors.New("invalid label:" + l)
			}
			t.Labels[s[0]] = s[1]
		}
	}
	return t, err
}

//Match compare event with topic
//If the match type is set to exact in long pulling request, only update request with exactly
//the same label of pulling request will match the request and will trigger an immediate return.
//
//If the match type is not set, it will be matched when pulling request labels is equal to
//update request labels or a subset of it.
func (t *Topic) Match(event *KVChangeEvent) bool {
	match := false
	if t.MatchType == common.PatternExact {
		if !util.IsEquivalentLabel(t.Labels, event.Labels) {
			return false
		}
	}
	if len(t.Labels) == 0 {
		return true
	}
	for k, v := range t.Labels {
		if event.Labels[k] != v {
			return false
		}
		match = true
	}
	return match
}

//Observer represents a client polling request
type Observer struct {
	UUID  string
	Event chan *KVChangeEvent
}
