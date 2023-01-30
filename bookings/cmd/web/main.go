package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/config"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/handlers"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/helpers"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/models"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var infoLog *log.Logger
var errorLog *log.Logger

// main is the main application function
func main() {
	err := run()
	if err != nil {
		log.Fatalf("Error setting up application: %q", err)
	}

	//	Start server
	fmt.Printf("Starting Web Server on port %s\n", portNumber)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("error starting server: %q", err)
	}
}

func run() error {
	// Configure application
	// change it to true when in production
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// Registering what we actually store in session
	gob.Register(models.Reservation{})
	// Creating a session instance
	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		return fmt.Errorf("error creating template cache: %w", err)
	}
	app.TemplateCache = tc
	app.UseCache = false
	render.NewTemplates(&app)
	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)
	helpers.NewHelpers(&app)

	return nil
}
