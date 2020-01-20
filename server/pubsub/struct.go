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
)

//KVChangeEvent is event between kie nodes, and broadcast by serf
type KVChangeEvent struct {
	Key      string
	Action   string //include: put,delete
	Labels   map[string]string
	DomainID string
	Project  string
}

//NewKVChangeEvent create a struct base on event payload
func NewKVChangeEvent(payload []byte) (*KVChangeEvent, error) {
	ke := &KVChangeEvent{}
	err := json.Unmarshal(payload, ke)
	return ke, err
}

//Topic can be subscribe
type Topic struct {
	Key          string            `json:"key,omitempty"`
	Labels       map[string]string `json:"-"`
	LabelsFormat string            `json:"labels,omitempty"`
	DomainID     string            `json:"domainID,omitempty"`
	Project      string            `json:"project,omitempty"`
}

//ParseTopicString parse topic string to topic struct
func ParseTopicString(s string) (*Topic, error) {
	t := &Topic{
		Labels: make(map[string]string),
	}
	err := json.Unmarshal([]byte(s), t)
	if err != nil {
		return nil, err
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
func (t *Topic) Match(event *KVChangeEvent) bool {
	match := false
	if t.Key != "" {
		if t.Key == event.Key {
			match = true
		}
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
	UUID      string
	RemoteIP  string
	UserAgent string
	Event     chan *KVChangeEvent
}
