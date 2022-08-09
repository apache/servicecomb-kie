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

package stringutil

import (
	"sort"
	"strings"
)

const (
	// LabelNone is the format string when the map is none
	LabelNone = "none"
)

// FormatMap format map to string
func FormatMap(m map[string]string) string {
	if len(m) == 0 {
		return LabelNone
	}
	sb := strings.Builder{}
	s := make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	sort.Strings(s)
	for i, k := range s {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(m[k])
		if i != (len(s) - 1) {
			sb.WriteString("::")
		}
	}
	return sb.String()
}
