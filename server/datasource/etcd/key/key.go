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

package key

import (
	"strconv"
	"strings"
)

const (
	split      = "/"
	keyKV      = "kvs"
	keyCounter = "counter"
	keyHistory = "kv-history"
	keyTrack   = "track"
	sync       = "sync"
	task       = "task"
)

func getSyncRootKey() string {
	return split + sync
}

func TaskKey(domain, project, timestamp string) string {
	return strings.Join([]string{getSyncRootKey(), task, domain, project, timestamp}, split)
}

func KV(domain, project, kvID string) string {
	return strings.Join([]string{keyKV, domain, project, kvID}, split)
}

func KVList(domain, project string) string {
	if len(project) == 0 {
		return strings.Join([]string{keyKV, domain, ""}, split)
	}
	return strings.Join([]string{keyKV, domain, project, ""}, split)
}

func Counter(name, domain string) string {
	return strings.Join([]string{keyCounter, domain, name}, split)
}

func His(domain, project, kvID string, updateRevision int64) string {
	return strings.Join([]string{keyHistory, domain, project, kvID,
		strconv.FormatInt(updateRevision, 10)}, split)
}

func HisList(domain, project, kvID string) string {
	return strings.Join([]string{keyHistory, domain, project, kvID, ""}, split)
}

func Track(domain, project, revision, sessionID string) string {
	return strings.Join([]string{keyTrack, domain, project, revision, sessionID}, split)
}

func TrackList(domain, project string) string {
	return strings.Join([]string{keyTrack, domain, project, ""}, split)
}
