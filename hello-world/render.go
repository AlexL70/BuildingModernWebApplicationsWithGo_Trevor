package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func renderTemplate(w http.ResponseWriter, tmpl string) {
	parsedTemplate, err := template.ParseFiles(fmt.Sprintf("./templates/%s", tmpl))
	if err != nil {
		fmt.Printf("error parsing template %s", err)
		return
	}
	err = parsedTemplate.Execute(w, nil)
	if err != nil {
		fmt.Printf("error parsing template %s", err)
	}
}
