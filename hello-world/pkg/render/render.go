package render

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// tc goes for template cache
var tc = make(map[string]*template.Template)

// RenderTemplate renders templates using html/template
func RenderTemplate(w http.ResponseWriter, t string) {
	var tmpl *template.Template
	var inMap bool
	var err error

	//	look for parsed template in cache
	tmpl, inMap = tc[t]
	if !inMap {
		log.Printf("Caching %q template", t)
		tmpl, err = addTemplateToCache(t)
		if err != nil {
			log.Printf("error adding template to cache: %s", err)
			return
		}
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("error executing template %s", err)
		return
	}
}

func addTemplateToCache(t string) (*template.Template, error) {
	templates := []string{
		fmt.Sprintf("./templates/%s", t),
		"./templates/base.layout.gohtml",
	}
	tmpl, err := template.ParseFiles(templates...)
	if err != nil {
		return nil, fmt.Errorf("errof parsing template %w", err)
	}
	tc[t] = tmpl
	return tmpl, nil
}
