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

func TestReadConfigFrom(t *testing.T) {
	testcases := []struct {
		name string
		toml string
	}{
		{
			name: "empty",
			toml: `version="unstable"`,
		},
		{
			name: "generator",
			toml: `
version="unstable"
generator="go"
`,
		},
		{
			name: "generators",
			toml: `
version="unstable"
generators=["go"]
`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := readConfigFrom([]byte(tc.toml))
			if err != nil {
				t.Fatalf("err must be nil, but got %v", err)
			}
			if c.Generator == "go" {
				t.Fatalf("Generator must be cleared, but got %v", c.Generator)
			}
			if len(c.Generators) != 1 || c.Generators[0] != "go" {
				t.Fatalf("Generators must be [go], but got %v", c.Generators)
			}
		})
	}
}
