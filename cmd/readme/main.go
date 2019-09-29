package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"libs.altipla.consulting/errors"
)

const tmplReadme = `
{{with .Logo}}{{.}}{{else}}# {{.Name}}{{end}}

[![GoDoc](https://godoc.org/libs.altipla.consulting/{{.Name}}?status.svg)](https://godoc.org/libs.altipla.consulting/{{.Name}})

{{.Doc}}


### Install

{{.Code}}go
import (
	"libs.altipla.consulting/{{.Name}}"
)
{{.Code}}

{{with .Extra}}{{.}}
{{end}}
### Contributing

You can make pull requests or create issues in GitHub. Any code you send should be formatted using {{.Code}}make gofmt{{.Code}}.


### Running tests

Run the tests:

{{.Code}}shell
make test
{{.Code}}


### License

[MIT License](../LICENSE)
`

func main() {
	if err := run(); err != nil {
		log.Fatal(errors.Stack(err))
	}
}

type readmeData struct {
	Name  string
	Doc   string
	Code  string
	Extra string
	Logo  string
}

func run() error {
	dirs, err := ioutil.ReadDir(".")
	if err != nil {
		return errors.Trace(err)
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		if strings.HasPrefix(dir.Name(), ".") || dir.Name() == "cmd" || dir.Name() == "infra" || dir.Name() == "protos" {
			continue
		}
		if len(os.Args) > 1 && dir.Name() != os.Args[1] {
			continue
		}

		log.WithField("package", dir.Name()).Info("Generate README file...")

		tmpl, err := template.New("readme").Parse(tmplReadme)
		if err != nil {
			return errors.Trace(err)
		}

		extra, err := ioutil.ReadFile(filepath.Join(dir.Name(), "_readme.mdtmpl"))
		if err != nil && !os.IsNotExist(err) {
			return errors.Trace(err)
		}
		logo, err := ioutil.ReadFile(filepath.Join(dir.Name(), "_logo.mdtmpl"))
		if err != nil && !os.IsNotExist(err) {
			return errors.Trace(err)
		}

		data := &readmeData{
			Name:  dir.Name(),
			Code:  "```",
			Extra: string(extra),
			Logo:  string(logo),
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, filepath.Join(dir.Name(), "doc.go"), nil, parser.ParseComments)
		if err != nil {
			return errors.Trace(err)
		}
		ast.Inspect(node, func(n ast.Node) bool {
			c, ok := n.(*ast.CommentGroup)
			if !ok {
				return true
			}

			var lines []string
			for _, l := range c.List {
				if l.Text == "//" {
					lines = append(lines, "")
					continue
				}

				lines = append(lines, l.Text[3:])
			}
			content := strings.Join(lines, "\n")

			prefix := fmt.Sprintf("Package %s ", dir.Name())
			if !strings.HasPrefix(content, prefix) {
				return true
			}

			data.Doc = fmt.Sprintf("Package `%s` %s", dir.Name(), content[len(prefix):])
			return true
		})
		if data.Doc == "" {
			return errors.Errorf("no doc found for package %q", dir.Name())
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return errors.Trace(err)
		}

		if err := ioutil.WriteFile(filepath.Join(dir.Name(), "README.md"), buf.Bytes(), 0700); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}
