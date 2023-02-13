package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/config"
	"github.com/AlexL70/BuildingModernWebApplicationsWithGo_Trevor/bookings/internal/driver"
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
	db, err := run()
	if err != nil {
		log.Fatalf("Error setting up application: %q", err)
	}
	defer db.SQL.Close()
	defer close(app.MailChan)
	log.Println("Starting mail listener...")
	listenForMail()

	// temporary code sending email message; to be deleted
	//msg := models.MailData{
	//	To:      "john@dow.ca",
	//	From:    "me@here.com",
	//	Subject: "Hi John!",
	//	Content: "Hello, <strong>world</strong>!",
	//}
	//app.MailChan <- msg

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

func run() (*driver.DB, error) {
	// Configure application
	// change it to true when in production
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// Registering what we actually store in session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	// Create mail channel
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	// Creating a session instance
	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	// connect to database
	log.Println("Connecting to the database")
	db, err := driver.ConnectSQL(os.Getenv("POSTGRESS_BOOKINGS_URL"))
	if err != nil {
		log.Fatal(fmt.Errorf("cannot connect to the database: %w. Dying", err))
	}
	log.Println("Connected to the DB!")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		return nil, fmt.Errorf("error creating template cache: %w", err)
	}
	app.TemplateCache = tc
	app.UseCache = false
	render.NewRenderer(&app)
	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	helpers.NewHelpers(&app)

	return db, nil
}
