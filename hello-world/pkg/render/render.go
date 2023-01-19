package render

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var tc map[string]*template.Template

func RenderTemplate(w http.ResponseWriter, tmpl string) {
	//	create a template cache
	tc, err := CreateTemplateCache()
	if err != nil {
		log.Fatalf("Error creating template cache: %q\n", err)
	}
	//	get requested template from cache
	t, ok := tc[tmpl]
	if !ok {
		log.Fatalf("Template called %q not found in cache.\n", tmpl)
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, nil)
	if err != nil {
		log.Printf("Error executing template: %q\n", err)
	}
	//	render the template
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Printf("Error writing parsed template to response writer: %q\n", err)
	}
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}
	//	get all files named *.page.gohtml from ./templates folder
	pages, err := filepath.Glob("./templates/*.page.gohtml")
	if err != nil {
		return myCache, err
	}
	matches, err := filepath.Glob("./templates/*.layout.gohtml")
	if err != nil {
		return myCache, err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return myCache, err
		}
		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.gohtml")
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}
	return myCache, nil
}
