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

	"encoding/json"

	"github.com/apache/servicecomb-kie/pkg/stringutil"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/go-chassis/openlog"
	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/hashicorp/serf/serf"
)

var once sync.Once
var bus *Bus

//const
const (
	EventKVChange             = "kv-chg"
	DefaultEventBatchSize     = 5000
	DefaultEventBatchInterval = 500 * time.Millisecond
)

var topics sync.Map

func Topics() *sync.Map {
	return &topics
}

//Bus is message bug
type Bus struct {
	agent *agent.Agent
}

//Init create serf agent
func Init() {
	once.Do(func() {
		ac := agent.DefaultConfig()
		sc := serf.DefaultConfig()
		setSerfAdvertiseAddr(sc, ac.BindAddr)
		if config.Configurations.ListenPeerAddr != "" {
			ac.BindAddr = config.Configurations.ListenPeerAddr
		}
		if config.Configurations.AdvertiseAddr != "" {
			ac.AdvertiseAddr = config.Configurations.AdvertiseAddr
			serfAdvertiseAddr := strings.Split(config.Configurations.AdvertiseAddr, ":")
			setSerfAdvertiseAddr(sc, serfAdvertiseAddr[0])
		}
		if config.Configurations.NodeName != "" {
			sc.NodeName = config.Configurations.NodeName
		}
		ac.UserEventSizeLimit = 512
		a, err := agent.Create(ac, sc, nil)
		if err != nil {
			openlog.Fatal("can not sync key value change events to other kie nodes:" + err.Error())
		}
		bus = &Bus{
			agent: a,
		}
		if config.Configurations.PeerAddr != "" {
			err := join([]string{config.Configurations.PeerAddr})
			if err != nil {
				openlog.Fatal("lost event message")
			} else {
				openlog.Info("join kie node:" + config.Configurations.PeerAddr)
			}
		}
	})
}

// set serf advertiseAddr value
func setSerfAdvertiseAddr(conf *serf.Config, advertiseAddr string) {
	conf.MemberlistConfig.AdvertiseAddr = advertiseAddr
}

//Start start serf agent
func Start() {
	err := bus.agent.Start()
	if err != nil {
		openlog.Fatal("can not sync key value change events to other kie nodes" + err.Error())
	}
	openlog.Info("kie message bus started")
	eh := &ClusterEventHandler{}
	bus.agent.RegisterEventHandler(eh)
}
func join(addresses []string) error {
	_, err := bus.agent.Join(addresses, false)
	if err != nil {
		return err
	}
	return nil
}

//Publish send event
func Publish(event *KVChangeEvent) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return bus.agent.UserEvent(EventKVChange, b, true)

}

//AddObserver observe key changes by (key or labels) or (key and labels)
func AddObserver(o *Observer, topic *Topic) (string, error) {
	topic.LabelsFormat = stringutil.FormatMap(topic.Labels)
	b, err := json.Marshal(topic)
	if err != nil {
		return "", err
	}
	t := string(b)
	observers, ok := topics.Load(t)
	if !ok {
		var observers = &sync.Map{}
		observers.Store(o.UUID, o)
		topics.Store(t, observers)
		openlog.Info("new topic:" + t)
		return t, nil
	}
	m := observers.(*sync.Map)
	m.Store(o.UUID, o)
	openlog.Debug("add new observer for topic:" + t)
	return t, nil
}
