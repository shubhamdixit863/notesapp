package app

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib" //use pgx in database/sql mode
)

// PostgreSQl configuration if not passed as env variables
const (
	host     = "ep-winter-limit-a57ruj96.us-east-2.aws.neon.tech" //127.0.0.1
	port     = 5432
	user     = "shubhamdixit863"
	password = "LQMlyi3r8hjT"
	dbname   = "rivaltrackdb"
)

var (
	err  error
	wait time.Duration
)

type App struct {
	Router   *mux.Router
	db       *pgx.Conn
	bindport string
	username string
	role     string
}

func (a *App) Initialize() {
	a.bindport = "80"

	//check if a different bind port was passed from the CLI
	//os.Setenv("PORT", "8080")
	tempport := os.Getenv("PORT")
	if tempport != "" {
		a.bindport = tempport
	}

	if len(os.Args) > 1 {
		s := os.Args[1]

		if _, err := strconv.ParseInt(s, 10, 64); err == nil {
			log.Printf("Using port %s", s)
			a.bindport = s
		}
	}
	connStr := "postgresql://shubhamdixit863:LQMlyi3r8hjT@ep-winter-limit-a57ruj96.us-east-2.aws.neon.tech/rivaltrackdb?sslmode=require"
	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close(context.Background())
	a.db = db
	//db, err = sql.Open("sqlite3", "db.sqlite3")
	if err != nil {
		log.Println("Invalid DB arguments, or github.com/lib/pq not installed")
		log.Fatal(err)
	}

	log.Println("Database connected successfully")

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	// setup static content route - strip ./assets/assets/[resource]
	// to keep /assets/[resource] as a route
	staticFileDirectory := http.Dir("./statics/")
	staticFileHandler := http.StripPrefix("/statics/", http.FileServer(staticFileDirectory))
	a.Router.PathPrefix("/statics/").Handler(staticFileHandler).Methods("GET")

	a.Router.HandleFunc("/login", a.loginHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/register", a.registerHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/notes", a.CreateNotesHandler).Methods("POST")

	log.Println("Routes established")

}

func (a *App) Run(addr string) {
	// Default to port 8080 if no address is provided
	if addr != "" {
		a.bindport = addr
	} else {
		a.bindport = "8080"
	}

	// Set up HTTP on Gorilla mux for a graceful shutdown
	srv := &http.Server{
		// Listen on the provided IP and port, or default to "0.0.0.0:8080" to allow external access
		Addr: "0.0.0.0:" + a.bindport,

		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      a.Router, // Assuming a.Router is already set up with the routes
	}

	// Start the HTTP listener in a goroutine as it's blocking
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
	}()

	// Set up a Ctrl-C trap to ensure a graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Shutting down HTTP service...")

	// Gracefully shutdown the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown failed: %v", err)
	}
	log.Println("HTTP service gracefully stopped")

	// Close database connections if applicable
	if a.db != nil {
		log.Println("Closing database connections")
		a.db.Close(context.Background())
	}

	log.Println("Shutting down the application")
	os.Exit(0)
}
