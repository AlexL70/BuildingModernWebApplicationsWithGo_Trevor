package config

import "html/template"

// AppConfig holds whole an application configuration
type AppConfig struct {
	TemplateCache map[string]*template.Template
}
