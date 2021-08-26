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
	"strings"
	"sync"
	"time"

	"github.com/go-chassis/openlog"
	"github.com/hashicorp/serf/serf"
)

//EventHandler handler serf custom event, it is singleton
type EventHandler struct {
	BatchSize          int
	BatchInterval      time.Duration
	Immediate          bool
	pendingEvents      sync.Map
	pendingEventsCount int
}

//HandleEvent send event to subscribers
func (h *EventHandler) HandleEvent(e serf.Event) {
	openlog.Debug("receive event:" + e.EventType().String())
	switch e.EventType().String() {
	case "user":
		if strings.Contains(e.String(), EventKVChange) {
			h.handleKVEvent(e)
		}
	}

}
func (h *EventHandler) RunFlushTask() {
	for {
		if h.pendingEventsCount >= h.BatchSize {
			h.fireEvents()
		}
		<-time.After(h.BatchInterval)
		h.fireEvents()
	}

}
func (h *EventHandler) handleKVEvent(e serf.Event) {
	ue := e.(serf.UserEvent)
	ke, err := NewKVChangeEvent(ue.Payload)
	if err != nil {
		openlog.Error("invalid json:" + string(ue.Payload))
	}
	openlog.Debug("kv event:" + ke.Key)
	if h.Immediate { //never retain event, not recommended
		h.FindTopicAndFire(ke)
	} else {
		h.mergeAndSave(ke)
	}

}
func (h *EventHandler) mergeAndSave(ke *KVChangeEvent) {
	id := ke.String()
	_, ok := h.pendingEvents.Load(id)
	if ok {
		openlog.Debug("ignore same event: " + id)
		return
	}
	h.pendingEvents.Store(id, ke)
	h.pendingEventsCount++
}
func (h *EventHandler) fireEvents() {
	h.pendingEvents.Range(func(key, value interface{}) bool {
		ke := value.(*KVChangeEvent)
		h.FindTopicAndFire(ke)
		h.pendingEvents.Delete(key)
		h.pendingEventsCount--
		return true
	})
}

func (h *EventHandler) FindTopicAndFire(ke *KVChangeEvent) {
	topics.Range(func(key, value interface{}) bool { //range all topics
		t, err := ParseTopicString(key.(string))
		if err != nil {
			openlog.Error("can not parse topic " + key.(string) + ": " + err.Error())
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
