package config

import "html/template"

// AppConfig holds whole an application configuration
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
}
