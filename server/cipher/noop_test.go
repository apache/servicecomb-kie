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

import "testing"

func TestNoop_Encrypt(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty", args{""}, "", false},
		{"123", args{"123"}, "123", false},
		{"nil", args{}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Noop{}
			got, err := n.Encrypt(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("Noop.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Noop.Encrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNoop_Decrypt(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty", args{""}, "", false},
		{"123", args{"123"}, "123", false},
		{"nil", args{}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Noop{}
			got, err := n.Decrypt(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("Noop.Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Noop.Decrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_namedNoop_Encrypt(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty", args{""}, "", false},
		{"123", args{"123"}, "123", false},
		{"nil", args{}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nn := &namedNoop{
				Name: "any",
			}
			got, err := nn.Encrypt(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("namedNoop.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("namedNoop.Encrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_namedNoop_Decrypt(t *testing.T) {

	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty", args{""}, "", false},
		{"123", args{"123"}, "123", false},
		{"nil", args{}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nn := &namedNoop{
				Name: "any",
			}
			got, err := nn.Decrypt(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("namedNoop.Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("namedNoop.Decrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}
