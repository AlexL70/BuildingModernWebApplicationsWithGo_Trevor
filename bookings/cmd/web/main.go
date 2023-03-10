package main

import (
	"encoding/gob"
	"flag"
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
	// Read flags
	inProduction := flag.Bool("production", true, "Application is in production")
	useCache := flag.Bool("cache", true, "Use template cache")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPassword := flag.String("dbpwd", "", "Database password")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database SSL settings (disable, prefer, require)")
	flag.Parse()

	// Configure application
	// change it to true when in production
	app.InProduction = *inProduction

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// Registering what we actually store in session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

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
	connStr := os.Getenv("POSTGRESS_BOOKINGS_URL")
	if connStr == "" {
		connStr = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
			*dbHost, *dbPort, *dbName, *dbUser, *dbPassword, *dbSSL)
	}
	db, err := driver.ConnectSQL(connStr)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot connect to the database: %w. Dying", err))
	}
	log.Println("Connected to the DB!")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		return nil, fmt.Errorf("error creating template cache: %w", err)
	}
	app.TemplateCache = tc
	app.UseCache = *useCache
	render.NewRenderer(&app)
	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	helpers.NewHelpers(&app)

	return db, nil
}
