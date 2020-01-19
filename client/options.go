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

package client

const (
	defaultProject = "default"
)

//GetOption is the functional option of client func
type GetOption func(*GetOptions)

//OpOption is the functional option of client func
type OpOption func(*OpOptions)

//GetOptions is the options of client func
type GetOptions struct {
	Labels  []map[string]string
	Project string
	Key     string
	Wait    string
}

//OpOptions is the options of client func
type OpOptions struct {
	Project string
}

//WithLabels query kv by labels
func WithLabels(l ...map[string]string) GetOption {
	return func(options *GetOptions) {
		for _, labels := range l {
			options.Labels = append(options.Labels, labels)
		}
	}
}

//WithGetProject query keys with certain project
func WithGetProject(project string) GetOption {
	return func(options *GetOptions) {
		options.Project = project
	}
}

//WithWait is for long polling,format is 1s,2m
func WithWait(duration string) GetOption {
	return func(options *GetOptions) {
		options.Wait = duration
	}
}

//WithKey query keys with certain key
func WithKey(k string) GetOption {
	return func(options *GetOptions) {
		options.Key = k
	}
}

//WithProject set project to param
func WithProject(project string) OpOption {
	return func(options *OpOptions) {
		options.Project = project
	}
}
