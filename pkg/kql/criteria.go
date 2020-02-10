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

package kql

var reservedInfix = []string{
	">=", "<=", "!=", "=", ">", "<", "AND", "OR",
}
var reservedPrefixSuffix = []string{
	"(", ")",
}

const (
	precedenceOr  = 1
	precedenceAnd = 2
)

type TreeNode struct {
	Expr      string
	Condition string
	Value     string
	children  []*TreeNode
}

//ParseKQL parse criteria for example "(a=b and b=a) or c=d"
//func ParseKQL(kql string) (mongo.Pipeline, error) {
//	strings.NewReader(kql)
//
//}
