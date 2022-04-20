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
		name       string
		cmd        protocCmd
		expectedV1 string
		expectedV2 string
	}{
		{
			name:       "basic",
			cmd:        protocCmd{Names: []string{"go"}},
			expectedV1: "protoc -I --go_out=import_path=:",
			expectedV2: "protoc -I --go_out= --go_opt=paths=source_relative",
		},
		{
			name:       "plugin",
			cmd:        protocCmd{Names: []string{"go"}, Plugins: []string{"grpc"}},
			expectedV1: "protoc -I --go_out=plugins=grpc,import_path=:",
			expectedV2: "protoc -I --go_out= --go_opt=paths=source_relative",
		},
		{
			name:       "use protoc-gen-go-grpc instead of plugins",
			cmd:        protocCmd{Names: []string{"go", "go-grpc"}},
			expectedV1: "protoc -I --go_out=import_path=: --go-grpc_out=import_path=:",
			expectedV2: "protoc -I --go_out= --go_opt=paths=source_relative --go-grpc_out= --go-grpc_opt=paths=source_relative",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name+"V1", func(t *testing.T) {
			cmd := &tc.cmd
			cmd.Version = 1

			s, err := cmd.mkcmd()
			if err != nil {
				t.Fatalf("err must be nil but %+v", err)
			}

			if s != tc.expectedV1 {
				t.Fatalf(`s must be %q, but %q`, tc.expectedV1, s)
			}
		})
		t.Run(tc.name+"V2", func(t *testing.T) {
			cmd := &tc.cmd
			cmd.Version = 2

			s, err := cmd.mkcmd()
			if err != nil {
				t.Fatalf("err must be nil but %+v", err)
			}

			if s != tc.expectedV2 {
				t.Fatalf(`s must be %q, but %q`, tc.expectedV2, s)
			}
		})
	}
}
