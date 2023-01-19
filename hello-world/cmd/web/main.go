package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/AlexL70/go-hello-world/pkg/config"
	"github.com/AlexL70/go-hello-world/pkg/handlers"
	"github.com/AlexL70/go-hello-world/pkg/render"
)

const portNumber = ":8080"

// main is the main application function
func main() {
	var app config.AppConfig
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatalf("error creating template cache: %q\n", err)
	}
	app.TemplateCache = tc

	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/about", handlers.About)

	fmt.Printf("Starting Web Server on port %s\n", portNumber)
	_ = http.ListenAndServe(portNumber, nil)
}
