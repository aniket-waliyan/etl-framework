package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Generator struct {
	Name string
}

func NewGenerator(name string) *Generator {
	return &Generator{Name: name}
}

func (g *Generator) Generate() error {
	pipelineDir := filepath.Join("pipelines", g.Name)
	if err := os.MkdirAll(pipelineDir, 0755); err != nil {
		return fmt.Errorf("failed to create pipeline directory: %w", err)
	}

	files := map[string]string{
		"etl.go":      etlTemplate,
		"config.yaml": configTemplate,
		"Dockerfile":  dockerfileTemplate,
		"README.md":   readmeTemplate,
		".env":        envTemplate,
	}

	for filename, tmpl := range files {
		if err := g.generateFile(pipelineDir, filename, tmpl); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filename, err)
		}
	}

	if err := g.generateRootEnvTemplate(); err != nil {
		return fmt.Errorf("failed to generate root .env.template: %w", err)
	}

	return nil
}

func (g *Generator) generateRootEnvTemplate() error {
	if _, err := os.Stat(".env.template"); os.IsNotExist(err) {
		return g.generateFile(".", ".env.template", envTemplate)
	}
	return nil
}

func (g *Generator) generateFile(dir, filename, tmpl string) error {
	filepath := filepath.Join(dir, filename)
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	data := struct {
		Name      string
		UpperName string
	}{
		Name:      g.Name,
		UpperName: strings.Title(g.Name),
	}

	t := template.Must(template.New(filename).Parse(tmpl))
	return t.Execute(f, data)
}
