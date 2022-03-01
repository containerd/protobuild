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
	"testing"
)

func TestRewrite(t *testing.T) {
	input := "//hello\npackage main\nfunc GetCpu(){}"
	expected := "//hello\npackage main\n\nfunc GetCPU() {}\n"

	fset := token.NewFileSet()
	n, err := parser.ParseFile(fset, "", input, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse: %s", err)
	}

	c := config{acronyms: []string{"Cpu"}}
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
