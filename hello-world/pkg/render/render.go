package render

import (
	"fmt"
	"html/template"
	"net/http"
)

// RenderTemplate renders templates using html/template
func RenderTemplate(w http.ResponseWriter, tmpl string) {
	parsedTemplate, err := template.ParseFiles(fmt.Sprintf("./templates/%s", tmpl), "./templates/base.layout.gohtml")
	if err != nil {
		fmt.Printf("error parsing template %s", err)
		return
	}
	err = parsedTemplate.Execute(w, nil)
	if err != nil {
		fmt.Printf("error parsing template %s", err)
	}
}
