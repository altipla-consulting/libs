package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/altipla-consulting/env"
	"github.com/altipla-consulting/errors"
)

type Group struct {
	root  string
	funcs template.FuncMap
	globs []string

	local *sync.Mutex
	tmpl  *template.Template
}

type PreloadOption func(g *Group)

func WithGlob(glob string) PreloadOption {
	return func(g *Group) {
		g.globs = append(g.globs, glob)
	}
}

func WithRoot(root string) PreloadOption {
	return func(g *Group) {
		g.root = root
	}
}

func WithFunc(name string, fn interface{}) PreloadOption {
	return func(g *Group) {
		g.funcs[name] = fn
	}
}

func newGroup(opts ...PreloadOption) *Group {
	g := &Group{
		funcs: make(template.FuncMap),
	}
	for name, fn := range stdfuncs {
		g.funcs[name] = fn
	}
	for _, opt := range opts {
		opt(g)
	}

	return g
}

func Preload(opts ...PreloadOption) (*Group, error) {
	g := newGroup(opts...)

	if len(g.globs) == 0 {
		return nil, errors.Errorf("no glob provided to the preload call")
	}

	if env.IsLocal() {
		g.local = new(sync.Mutex)
		if g.root != "" {
			for i, glob := range g.globs {
				g.globs[i] = filepath.Join(g.root, glob)
			}
		}
	}

	if err := g.load(); err != nil {
		return nil, errors.Trace(err)
	}

	return g, nil
}

func NewFromReader(r io.Reader, opts ...PreloadOption) (*Group, error) {
	g := newGroup(opts...)

	g.tmpl = template.New("root").Funcs(g.funcs)

	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	g.tmpl, err = g.tmpl.Parse(string(content))
	if err != nil {
		return nil, errors.Trace(err)
	}

	return g, nil
}

func (g *Group) load() error {
	// If there is no file to load this function does nothing.
	if len(g.globs) == 0 {
		return nil
	}

	if env.IsLocal() {
		// In development reload every single time protecting of common memory
		// issues with the lock.
		g.local.Lock()
		defer g.local.Unlock()
	} else if g.tmpl != nil {
		// In production load the templates only the first time when calling Preload().
		return nil
	}

	g.tmpl = template.New("root").Funcs(g.funcs)

	for _, glob := range g.globs {
		if _, err := g.tmpl.ParseGlob(glob); err != nil {
			return fmt.Errorf("glob %q: %w", glob, err)
		}
	}

	return nil
}

func (g *Group) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	if err := g.load(); err != nil {
		return errors.Trace(err)
	}

	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return errors.Trace(err)
	}

	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	fmt.Fprint(w, buf.String())

	return nil
}

func (g *Group) Execute(w io.Writer, data interface{}) error {
	if err := g.load(); err != nil {
		return errors.Trace(err)
	}

	var buf bytes.Buffer
	if err := g.tmpl.Execute(&buf, data); err != nil {
		return errors.Trace(err)
	}

	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	fmt.Fprint(w, buf.String())

	return nil
}
