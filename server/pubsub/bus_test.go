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

package pubsub_test

import (
	"testing"

	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/pubsub"
	uuid "github.com/satori/go.uuid"
)

func TestInit(t *testing.T) {
	config.Configurations = &config.Config{}
	pubsub.Init()
	pubsub.Start()

	o := &pubsub.Observer{
		UUID:  uuid.NewV4().String(),
		Event: make(chan *pubsub.KVChangeEvent, 1),
	}
	_ = pubsub.ObserveOnce(o, &pubsub.Topic{
		Key:      "some_key",
		Project:  "1",
		DomainID: "2",
		Labels: map[string]string{
			"a": "b",
			"c": "d",
		},
	})
	_ = pubsub.Publish(&pubsub.KVChangeEvent{
		Key:    "some_key",
		Action: "put",
		Labels: map[string]string{
			"a": "b",
			"c": "d",
		},
		Project:  "1",
		DomainID: "2",
	})
	e := <-o.Event
	t.Log(e.Key)
}
