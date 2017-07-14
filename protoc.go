package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"text/template"
)

var (
	tmpl = template.Must(template.New("protoc").Parse(`protoc -I
	{{- range $index, $include := .Includes -}}
		{{if $index}}:{{end -}}
			{{.}}
		{{- end }} --
	{{- .Name -}}_out={{if .Plugins}}plugins={{- range $index, $plugin := .Plugins -}}
		{{- if $index}}+{{end}}
		{{- $plugin}}
	{{- end -}}
	,{{- end -}}import_path={{.ImportPath}}
	{{- range $proto, $gopkg := .PackageMap -}},M
		{{- $proto}}={{$gopkg -}}
	{{- end -}}
	:{{- .OutputDir }}
	{{- range .Files}} {{.}}{{end -}}
`))
)

// protocParams defines inputs to a protoc command string.
type protocCmd struct {
	Name       string // backend name
	Includes   []string
	Plugins    []string
	ImportPath string
	PackageMap map[string]string
	Files      []string
	OutputDir  string
}

func (p *protocCmd) mkcmd() (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, p); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (p *protocCmd) run() error {
	arg, err := p.mkcmd()
	if err != nil {
		log.Fatalln(err)
	}

	// pass to sh -c so we don't need to re-split here.
	args := []string{"-c", arg}
	cmd := exec.Command("sh", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
