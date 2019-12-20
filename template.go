package telebot

import (
	"bytes"
	"io/ioutil"
	"path"
	"strings"
	"text/template"

	"github.com/aymerick/raymond"
)

// Template implements interface to parse with templates.
// Implement initializes a template engine (make new structure point, etc).
// Execute is function, which it is parsing template using the selected key.
type Template interface {
	Implement() error
	Execute(*bytes.Buffer, string, interface{}) error
}

// TemplateText implements a template interface using the text/template library.
type TemplateText struct {
	tmpl *template.Template
	Dir  string
}

// Execute parses template.
func (t *TemplateText) Execute(buf *bytes.Buffer, key string, arg interface{}) error {
	return t.tmpl.ExecuteTemplate(buf, key+".tmpl", arg)
}

// Implement initializes a template engine.
func (t *TemplateText) Implement() error {
	tmpl, err := template.New("data").
		Funcs(TemplateFuncMap).
		ParseGlob(t.Dir + "/*.tmpl")
	if err != nil {
		return err
	}

	t.tmpl = tmpl
	return nil
}

// TemplateHandlebars implements a template interface using the aymerick/raymond library.
type TemplateHandlebars struct {
	handlers map[string]*raymond.Template
	Dir      string
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

// Implement initializes a template engine.
func (t *TemplateHandlebars) Implement() error {
	files, err := ioutil.ReadDir(t.Dir)
	if err != nil {
		return err
	}

	t.handlers = make(map[string]*raymond.Template)
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
		t.handlers[strings.TrimSuffix(file.Name(), ".tmpl")] = tmpl
	}
	return nil
}
