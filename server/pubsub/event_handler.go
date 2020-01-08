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
	"github.com/go-mesh/openlogging"
	"github.com/hashicorp/serf/serf"
	"strings"
)

//EventHandler handler serf custom event
type EventHandler struct {
}

//HandleEvent send event to subscribers
func (h *EventHandler) HandleEvent(e serf.Event) {
	openlogging.Info("receive event:" + e.EventType().String())
	switch e.EventType().String() {
	case "user":
		if strings.Contains(e.String(), EventKVChange) {
			handleKVEvent(e)
		}
	}

}

func handleKVEvent(e serf.Event) {
	ue := e.(serf.UserEvent)
	ke, err := NewKVChangeEvent(ue.Payload)
	if err != nil {
		openlogging.Error("invalid json:" + string(ue.Payload))
	}
	openlogging.Debug("kv event:" + ke.Key)
	topics.Range(func(key, value interface{}) bool { //range all topics
		t, err := ParseTopicString(key.(string))
		if err != nil {
			openlogging.Error("can not parse topic:" + key.(string))
			return true
		}
		if t.Match(ke) {
			fireEvent(value, ke)
		}
		return true
	})
}

func fireEvent(value interface{}, ke *KVChangeEvent) {
	observers := value.(map[string]*Observer)
	mutexObservers.Lock()
	defer mutexObservers.Unlock()
	for k, v := range observers {
		v.Event <- ke
		delete(observers, k)
	}
}
