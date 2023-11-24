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

package common

import "time"

// match mode
const (
	QueryParamQ            = "q"
	QueryByLabelsCon       = "&"
	QueryParamWait         = "wait"
	QueryParamRev          = "revision"
	QueryParamMatch        = "match"
	QueryParamKey          = "key"
	QueryParamValue        = "value"
	QueryParamLabel        = "label"
	QueryParamStatus       = "status"
	QueryParamOffset       = "offset"
	QueryParamLimit        = "limit"
	PathParamKVID          = "kv_id"
	PathParameterProject   = "project"
	QueryParamSessionID    = "sessionId"
	QueryParamSessionGroup = "sessionGroup"
	QueryParamIP           = "ip"
	QueryParamURLPath      = "urlPath"
	QueryParamUserAgent    = "userAgent"
	QueryParamOverride     = "override"
)

// http headers
const (
	HeaderDepth       = "X-Depth"
	HeaderRevision    = "X-Kie-Revision"
	HeaderContentType = "Content-Type"
	HeaderAccept      = "Accept"
)

// ContentType
const (
	ContentTypeText = "application/text"
	ContentTypeJSON = "application/json"
	ContentTypeYaml = "text/yaml"
)

// const of server
const (
	PatternExact            = "exact"
	StatusEnabled           = "enabled"
	StatusDisabled          = "disabled"
	MsgDomainMustNotBeEmpty = "domain must not be empty"
	MsgDeleteKVFailed       = "delete kv failed"
	MsgIllegalLabels        = "label value can not be empty, " +
		"label can not be duplicated, please check query parameters"
	MsgIllegalDepth    = "X-Depth must be number"
	MsgInvalidWait     = "wait param should be formed with number and time unit like 5s,100ms, and less than 5m"
	MsgInvalidRev      = "revision param should be formed with number greater than 0"
	RespBodyContextKey = "responseBody"

	MaxWait = 5 * time.Minute
)

// all msg server returns
const (
	MsgDBError = "database operation error"
)
