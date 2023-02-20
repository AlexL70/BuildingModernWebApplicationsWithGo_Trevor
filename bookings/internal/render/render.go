package render

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/config"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/models"
	"github.com/justinas/nosurf"
)

var functions = template.FuncMap{
	"humanDate": HumanDate,
}

var app *config.AppConfig
var pathToTemplates string = "./templates"

// NewRenderer set the config for render package
func NewRenderer(ac *config.AppConfig) {
	app = ac
}

// HumanDate returns time.Time as a string formatten as "YYYY-MM-DD"
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// AddDefaultData adds data for all templates
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}
	td.CSRFToken = nosurf.Token(r)
	return td
}

// Template renders templates using html/template
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	if !app.UseCache {
		log.Println("Reloading templates' cache...")
		tc, err := CreateTemplateCache()
		if err != nil {
			err = fmt.Errorf("error caching templates: %w", err)
			log.Println(err)
			return err
		}
		app.TemplateCache = tc
	}
	//	get requested template from cache
	t, ok := app.TemplateCache[tmpl]
	if !ok {
		err := fmt.Errorf("template called %q not found in cache", tmpl)
		log.Println(err)
		return err
	}
	buf := new(bytes.Buffer)
	fmt.Println("Adding default data...")
	td = AddDefaultData(td, r)
	fmt.Println("Error:", td.Error)
	fmt.Println("Warning:", td.Warning)
	fmt.Println("Flash:", td.Flash)
	err := t.Execute(buf, td)
	if err != nil {
		err := fmt.Errorf("error executing template: %w", err)
		log.Println(err)
		return err
	}
	//	render the template
	_, err = buf.WriteTo(w)
	if err != nil {
		err := fmt.Errorf("error writing parsed template to response writer: %w", err)
		log.Println(err)
		return err
	}
	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}
	//	get all files named *.page.gohtml from ./templates folder
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.gohtml", pathToTemplates))
	if err != nil {
		return myCache, err
	}
	matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.gohtml", pathToTemplates))
	if err != nil {
		return myCache, err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}
		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.gohtml", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}
	return myCache, nil
}
