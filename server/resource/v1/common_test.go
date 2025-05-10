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

package v1

import (
	"errors"
	"github.com/apache/servicecomb-kie/server/pubsub"
	_ "github.com/apache/servicecomb-kie/test"
	"net/http"
	"testing"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
)

func TestGetLabels(t *testing.T) {
	r, err := http.NewRequest("GET",
		"/kv?q=app:mall+service:payment&q=app:mall+service:payment+version:1.0.0",
		nil)
	assert.NoError(t, err)
	c, err := ReadLabelCombinations(restful.NewRequest(r))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(c))

	r, err = http.NewRequest("GET",
		"/kv",
		nil)
	assert.NoError(t, err)
	c, err = ReadLabelCombinations(restful.NewRequest(r))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(c))
}

func Test_splitWaitInterval(t *testing.T) {
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name                          string
		args                          args
		wantNormalInterval            time.Duration
		wantTimeoutProtectionInterval time.Duration
	}{
		{
			name: "regular interval",
			args: args{
				d: 5 * time.Second,
			},
			wantNormalInterval:            4 * time.Second,
			wantTimeoutProtectionInterval: 1 * time.Second,
		},
		{
			name: "interval very big, origin timeout protect bigger than the max value",
			args: args{
				d: 5000 * time.Second,
			},
			wantNormalInterval:            4998 * time.Second,
			wantTimeoutProtectionInterval: 2 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNormalInterval, gotTimeoutProtectionInterval := splitWaitInterval(tt.args.d)
			assert.Equalf(t, tt.wantNormalInterval, gotNormalInterval, "splitWaitInterval(%v)", tt.args.d)
			assert.Equalf(t, tt.wantTimeoutProtectionInterval, gotTimeoutProtectionInterval, "splitWaitInterval(%v)", tt.args.d)
		})
	}
}

func getObserverThatHappenedEventAfter10ms() *pubsub.Observer {
	observer, _ := NewObserver()
	go func() {
		time.Sleep(10 * time.Millisecond)
		observer.Event <- &pubsub.KVChangeEvent{}
	}()
	return observer
}

func Test_waitObserverEventHappened(t *testing.T) {
	// timeout
	assert.False(t, waitObserverEventHappened(getObserverThatHappenedEventAfter10ms(), 5*time.Millisecond))

	// not timeout
	assert.True(t, waitObserverEventHappened(getObserverThatHappenedEventAfter10ms(), 50*time.Millisecond))

	// zero interval
	assert.False(t, waitObserverEventHappened(getObserverThatHappenedEventAfter10ms(), 0*time.Millisecond))
}

func Test_isDeadlineExceededErr(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "DeadlineExceededErr",
			args: args{
				err: errors.New("error is context deadline exceeded"),
			},
			want: true,
		},
		{
			name: "no DeadlineExceededErr",
			args: args{
				err: errors.New("error is unknown"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, isDeadlineExceededErr(tt.args.err), "isDeadlineExceededErr(%v)", tt.args.err)
		})
	}
}
