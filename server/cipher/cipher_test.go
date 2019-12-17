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
	"reflect"
	"testing"
)

func TestLookup(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"noop", args{"noop", "123"}, "123"},
		{"namedNoop", args{"not_implemented", "123"}, "123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCipher := Lookup(tt.args.name)
			expect, _ := gotCipher.Encrypt(tt.args.value)
			if !reflect.DeepEqual(expect, tt.want) {
				t.Errorf("Lookup() = %v, want %v", expect, tt.want)
			}

			expect, _ = gotCipher.Decrypt(tt.args.value)
			if !reflect.DeepEqual(expect, tt.want) {
				t.Errorf("Lookup() = %v, want %v", expect, tt.want)
			}
		})
	}
}

type testCipher struct{}

func (*testCipher) Encrypt(src string) (string, error) {
	return src, nil
}

func (*testCipher) Decrypt(src string) (string, error) {
	return src, nil
}

func TestRegister(t *testing.T) {
	test := &testCipher{}
	noop2 := &namedNoop{}
	Register("test", test)
	Register("noop2", noop2)

	act := Lookup("test")
	if !reflect.DeepEqual(test, act) {
		t.Errorf("Register() = %v, want %v", test, act)
	}
	act = Lookup("noop2")
	if !reflect.DeepEqual(noop2, act) {
		t.Errorf("Register() = %v, want %v", noop2, act)
	}
}
