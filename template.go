package telebot

import (
	"bytes"
	"errors"
	"io/ioutil"
	"path"
	"strings"
	"text/template"

	"github.com/aymerick/raymond"
)

var (
	ErrTemplateIsNil    = errors.New("telebot: template is not initalized")
	ErrTemplateEmptyDir = errors.New("telebot: template dir is empty")
)

// TemplateFuncMap is pre-defined functions that can be used in your text/template templates.
// It's better to not edit this global map.
var TemplateFuncMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },

	// Escapes double-quotes. Useful in json templates.
	"jsq": func(s string) string {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		return s
	},

	// String functions
	"title":  strings.ToTitle,
	"repeat": strings.Repeat,
}

// Template implements interface to parse with templates.
// Always call New() before manipulating the template.
type Template interface {
	// New initializes a template engine and returns copy of itself with passed params.
	New(string) Template
	// Parse parses given value and stores it by passed key.
	Parse(string, string) error
	// ParseGlob parses all directory files in passed Dir using glob pattern.
	ParseGlob() error
	// Execute executes template using the selected key.
	Execute(*bytes.Buffer, string, interface{}) error
}

// TemplateText implements a template interface using the text/template library.
type TemplateText struct {
	tmpl *template.Template

	Dir        string
	Funcs      template.FuncMap
	DelimLeft  string
	DelimRight string
}

// New returns initialized template.
func (t *TemplateText) New(name string) Template {
	if t.DelimLeft == "" || t.DelimRight == "" {
		t.DelimLeft, t.DelimRight = "{{", "}}"
	}

	tmpl := template.New(name).
		Funcs(TemplateFuncMap).
		Funcs(t.Funcs).
		Delims(t.DelimLeft, t.DelimRight)

	cpy := *t
	cpy.tmpl = tmpl
	return &cpy
}

// Parse parses template and stores it by passed key.
func (t *TemplateText) Parse(key, value string) error {
	if t.tmpl == nil {
		return ErrTemplateIsNil
	}

	tmpl, err := t.tmpl.New(key).Parse(value)
	if err != nil {
		return err
	}

	t.tmpl = tmpl
	return err
}

// ParseGlob parses all directory templates.
func (t *TemplateText) ParseGlob() error {
	if t.tmpl == nil {
		return ErrTemplateIsNil
	}
	if t.Dir == "" {
		return ErrTemplateEmptyDir
	}

	tmpl, err := t.tmpl.ParseGlob(t.Dir + "/*.tmpl")
	if err != nil {
		return err
	}

	t.tmpl = tmpl
	return nil
}

// Execute parses template.
func (t *TemplateText) Execute(buf *bytes.Buffer, key string, arg interface{}) error {
	return t.tmpl.ExecuteTemplate(buf, key, arg)
}

// TemplateHandlebars implements a template interface using the aymerick/raymond library.
type TemplateHandlebars struct {
	handlers map[string]*raymond.Template

	Name string
	Dir  string
}

// New returns initialized template.
func (t *TemplateHandlebars) New(name string) Template {
	cpy := *t
	cpy.handlers = make(map[string]*raymond.Template)
	cpy.Name = name
	return &cpy
}

// ParseGlob parses all directory templates.
func (t *TemplateHandlebars) Parse(key, value string) error {
	if t.handlers == nil {
		return ErrTemplateIsNil
	}

	tmpl, err := raymond.Parse(value)
	if err != nil {
		return err
	}

	t.handlers[key] = tmpl
	return nil
}

// New returns initialized template.
func (t *TemplateHandlebars) ParseGlob() error {
	if t.handlers == nil {
		return ErrTemplateIsNil
	}
	if t.Dir == "" {
		return ErrTemplateEmptyDir
	}

	files, err := ioutil.ReadDir(t.Dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".tmpl") {
			continue
		}

		tmpl, err := raymond.ParseFile(path.Join(t.Dir, file.Name()))
		if err != nil {
			return err
		}

		t.handlers[file.Name()] = tmpl
	}

	return nil
}

// Execute parses template.
func (t *TemplateHandlebars) Execute(buf *bytes.Buffer, key string, arg interface{}) error {
	tmpl, ok := t.handlers[key]
	if !ok {
		return nil
	}

	result, err := tmpl.Exec(arg)
	if err != nil {
		return err
	}

	_, err = buf.WriteString(result)
	return err
}
