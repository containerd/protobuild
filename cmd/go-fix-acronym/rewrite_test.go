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

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func testRewrite(t *testing.T, input, expected string, c config) {
	fset := token.NewFileSet()
	n, err := parser.ParseFile(fset, "", input, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse: %s", err)
	}

	p, err := compilePattern(c)
	if err != nil {
		t.Fatalf("failed to compile: %s", err)
	}

	rewriteNode(p, n)

	out := &bytes.Buffer{}
	format.Node(out, fset, n)
	if out.String() != expected {
		t.Fatalf("expected %q, but got %q", expected, out)
	}
}

func TestRewrite(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		expected string
		c        config
	}{
		{
			name:     "Simple",
			c:        config{acronyms: []string{"Cpu"}},
			input:    "//hello\npackage main\nfunc GetCpu(){}",
			expected: "//hello\npackage main\n\nfunc GetCPU() {}",
		},
		{
			name: "Multiple matches",
			c:    config{acronyms: []string{"Cpu"}},
			input: `package main

			func GetCpuFromCpuList() {}`,
			expected: `package main

			func GetCPUFromCPUList() {}`,
		},
		{
			name: "Submatches",
			c:    config{acronyms: []string{"Runtime(Ns)"}},
			input: `package main

			func KernelTime_100Ns()            {}
			func RuntimeNsAndNsAndSomeSuffix() {}`,
			expected: `package main

			func KernelTime_100Ns()            {}
			func RuntimeNSAndNsAndSomeSuffix() {}`,
		},
		{
			name: "Multiple submatches",
			c:    config{acronyms: []string{"(Id|Vm)$", "[a-z](Ns)$"}},
			input: `package main

			func Vm()         {}
			func Time_100Ns() {}
			func RuntimeNs()  {}
			`,
			expected: `package main

			func VM()         {}
			func Time_100Ns() {}
			func RuntimeNS()  {}`,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testRewrite(t, tc.input, strings.ReplaceAll(tc.expected, "\t", "")+"\n", tc.c)
		})
	}
}
