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
	"go/ast"
	"regexp"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func compilePattern(c config) (*regexp.Regexp, error) {
	return regexp.Compile(strings.Join(c.acronyms, "|"))
}

func rewriteNode(pattern *regexp.Regexp, node ast.Node) {
	astutil.Apply(
		node,
		func(c *astutil.Cursor) bool {
			node := c.Node()
			ident, ok := node.(*ast.Ident)
			if !ok {
				return true
			}
			name := pattern.ReplaceAllFunc([]byte(ident.Name), func(b []byte) []byte {
				return []byte(strings.ToUpper(string(b)))
			})
			ident.Name = string(name)
			return false
		},
		nil,
	)
}
