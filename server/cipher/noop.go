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

package cipher

import (
	"fmt"
	"github.com/go-mesh/openlogging"
)

// Noop is none implement
type Noop struct {
}

// Encrypt implement
func (*Noop) Encrypt(src string) (string, error) {
	return src, nil
}

// Decrypt implement
func (*Noop) Decrypt(src string) (string, error) {
	return src, nil
}

type namedNoop struct {
	Name string
}

func (nn *namedNoop) Encrypt(src string) (string, error) {
	openlogging.Warn(fmt.Sprintf("security name [%s] not implemented.", nn.Name))
	return src, nil
}

func (nn *namedNoop) Decrypt(src string) (string, error) {
	openlogging.Warn(fmt.Sprintf("security name [%s] not implemented.", nn.Name))
	return src, nil
}
