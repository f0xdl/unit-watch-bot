package templates

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sync"
	"text/template"
)

type TemplateService struct {
	templates map[string]*template.Template
	mu        sync.RWMutex
}

func New(basePath string) (*TemplateService, error) {
	ts := &TemplateService{}
	if err := ts.LoadTemplates(basePath); err != nil {
		return nil, err
	}
	return ts, nil
}

// LoadTemplates Load from basePath
func (ts *TemplateService) LoadTemplates(basePath string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.templates = map[string]*template.Template{}

	pattern := filepath.Join(basePath, "*.tmpl")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("error in read templates %w", err)
	}

	for _, file := range files {
		name := filepath.Base(file)
		tmpl, err := template.ParseFiles(file)
		if err != nil {
			return fmt.Errorf("error in parse template '%s': %w", name, err)
		}
		ts.templates[name] = tmpl
	}
	return nil
}

func (ts *TemplateService) Render(name string, data interface{}) (string, error) {
	ts.mu.RLock()
	tmpl, ok := ts.templates[name]
	ts.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error rendering %s: %w", name, err)
	}
	return buf.String(), nil
}
