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
	"encoding/json"
	"testing"

	"github.com/apache/servicecomb-kie/server/pubsub"
)

func TestTopic_String(t *testing.T) {
	topic := &pubsub.Topic{
		Labels: map[string]string{
			"a": "b",
			"c": "d",
		},
	}
	t.Log(topic)
	b, _ := json.Marshal(topic)
	t.Log(string(b))
	topic = &pubsub.Topic{
		Labels: map[string]string{
			"a": "b",
			"c": "d",
		},
	}
	t.Log(topic)
	b, _ = json.Marshal(topic)
	t.Log(string(b))
	topic = &pubsub.Topic{}
	t.Log(topic)
	b, _ = json.Marshal(topic)
	t.Log(string(b))

	mock := []byte(`{"labels":"a=b::c=d","domainID":"2","project":"1"}`)
	topic, _ = pubsub.ParseTopic(string(mock))
	t.Log(topic)
}
