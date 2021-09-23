/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import "testing"

func TestMkcmd(t *testing.T) {
	testcases := []struct {
		name     string
		cmd      protocCmd
		expected string
	}{
		{
			name:     "basic",
			cmd:      protocCmd{Names: []string{"go"}},
			expected: "protoc -I --go_out=import_path=:",
		},
		{
			name:     "plugin",
			cmd:      protocCmd{Names: []string{"go"}, Plugins: []string{"grpc"}},
			expected: "protoc -I --go_out=plugins=grpc,import_path=:",
		},
		{
			name:     "use protoc-gen-go-grpc instead of plugins",
			cmd:      protocCmd{Names: []string{"go", "go-grpc"}},
			expected: "protoc -I --go_out=import_path= --go-grpc_out=import_path=:",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &tc.cmd

			s, err := cmd.mkcmd()
			if err != nil {
				t.Fatalf("err must be nil but %+v", err)
			}

			if s != tc.expected {
				t.Fatalf(`s must be %q, but %q`, tc.expected, s)
			}
		})
	}
}
