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

package validator

import "github.com/go-chassis/foundation/validator"

const (
	key                   = "key"
	keyRegex              = `^[a-zA-Z0-9._:-]+$`
	commonNameRegexString = `^[a-zA-Z0-9]*$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]*[a-zA-Z0-9]$`
	labelKeyRegexString   = `^[a-zA-Z0-9]{1,32}$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]{1,30}[a-zA-Z0-9]$`
	labelValueRegexString = `^[a-zA-Z0-9]{0,160}$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]{0,158}[a-zA-Z0-9]$`
	getKeyRegexString     = `^[a-zA-Z0-9._:-]*$|^beginWith\([a-zA-Z0-9._:-]*\)$|^wildcard\([a-zA-Z0-9*._:-]*\)$`
	asciiRegexString      = `^[\x00-\x7F]*$`
	allCharString         = `.*`
)

// custom validate rules
// please use different tag names from third party tags
var customRules = []*validator.RegexValidateRule{
	validator.NewRegexRule(key, keyRegex),
	validator.NewRegexRule("getKey", getKeyRegexString),
	validator.NewRegexRule("commonName", commonNameRegexString),
	validator.NewRegexRule("valueType", `^$|^(ini|json|text|yaml|properties|xml)$`),
	validator.NewRegexRule("kvStatus", `^$|^(enabled|disabled)$`),
	validator.NewRegexRule("value", allCharString), //ASCII, 128k
	validator.NewRegexRule("labelK", labelKeyRegexString),
	validator.NewRegexRule("labelV", labelValueRegexString),
	validator.NewRegexRule("check", asciiRegexString), //ASCII, 1M
}

func Init() error {
	return validator.RegisterRegexRules(customRules)
}
