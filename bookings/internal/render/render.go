package render

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/config"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/models"
	"github.com/justinas/nosurf"
)

var app *config.AppConfig

// set the config for render package
func NewTemplates(ac *config.AppConfig) {
	app = ac
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.CSRFToken = nosurf.Token(r)
	return td
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) {
	if !app.UseCache {
		log.Println("Reloading templates' cache...")
		tc, err := CreateTemplateCache()
		if err != nil {
			log.Printf("Error caching templates: %q\n", err)
			return
		}
		app.TemplateCache = tc
	}
	//	get requested template from cache
	t, ok := app.TemplateCache[tmpl]
	if !ok {
		log.Fatalf("Template called %q not found in cache.\n", tmpl)
	}
	buf := new(bytes.Buffer)
	td = AddDefaultData(td, r)
	err := t.Execute(buf, td)
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
