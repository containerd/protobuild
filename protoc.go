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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	tmpl = template.Must(template.New("protoc").Parse(`protoc -I
	{{- range $index, $include := .Includes -}}
		{{if $index}}` + string(filepath.ListSeparator) + `{{end -}}
			{{.}}
	{{- end -}}
	{{- if .Descriptors}} --include_imports --descriptor_set_out={{.Descriptors}}{{- end -}}

	{{ if lt .Version 2 }}
		{{- range $index, $name := .Names }} --{{- $name -}}_out={{- $.GoOutV1 }}{{- end -}}
	{{- else -}}
		{{- range $index, $name := .Names }} --{{- $name -}}_out={{- $.GoOutV2 }}{{- end -}}

		{{- range $proto, $gopkg := .PackageMap }} --go_opt=M
			{{- $proto}}={{$gopkg -}}
		{{- end -}}
	{{- end -}}

	{{- range .Files}} {{.}}{{end -}}
`))
)

// protocParams defines inputs to a protoc command string.
type protocCmd struct {
	Names       []string
	Includes    []string
	Plugins     []string
	Descriptors string
	ImportPath  string
	PackageMap  map[string]string
	Files       []string
	OutputDir   string
	// Version is Protobuild's version.
	Version int
}

func (p *protocCmd) mkcmd() (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, p); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GoOutV1 returns the parameter for --go_out= for protoc-gen-go < 1.4.0.
// Note that plugins and import_path are no longer supported by
// newer protoc-gen-go versions.
func (p *protocCmd) GoOutV1() string {
	var result string
	if len(p.Plugins) > 0 {
		result += "plugins=" + strings.Join(p.Plugins, "+") + ","
	}
	result += "import_path=" + p.ImportPath

	for proto, pkg := range p.PackageMap {
		result += fmt.Sprintf(",M%s=%s", proto, pkg)
	}
	result += ":" + p.OutputDir

	return result
}

// GoOutV2 returns the parameter for --go_out= for protoc-gen-go >= 1.4.0.
func (p *protocCmd) GoOutV2() string {
	return p.OutputDir
}

func (p *protocCmd) run() error {
	arg, err := p.mkcmd()
	if err != nil {
		return err
	}

	// pass to sh -c so we don't need to re-split here.
	args := []string{shArg, arg}
	cmd := exec.Command(shCmd, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
