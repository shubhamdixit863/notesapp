package app

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib" //use pgx in database/sql mode
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

func NewApp() *App {
	return &App{}
}

func (a *App) Initialize() {

	connStr := "postgresql://shubhamdixit863:LQMlyi3r8hjT@ep-winter-limit-a57ruj96.us-east-2.aws.neon.tech/rivaltrackdb?sslmode=require"
	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	a.db = db

	log.Println("Database connected successfully")

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	// setup static content route - strip ./assets/assets/[resource]
	// to keep /assets/[resource] as a route
	a.Router.HandleFunc("/health", a.healthHandler)

	a.Router.HandleFunc("/login", a.loginHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/register", a.registerHandler).Methods("POST", "GET")
	a.Router.Handle("/notes", JWTAuthMiddleware(http.HandlerFunc(a.CreateNotesHandler))).Methods("POST")
	a.Router.Handle("/notes/{id}", JWTAuthMiddleware(http.HandlerFunc(a.GetNoteById))).Methods("GET")
	a.Router.Handle("/notes", JWTAuthMiddleware(http.HandlerFunc(a.GetNotesHandler))).Methods("GET")
	a.Router.Handle("/notes/{id}", JWTAuthMiddleware(http.HandlerFunc(a.UpdateNotesHandler))).Methods("PUT")
	a.Router.Handle("/notes/{id}", JWTAuthMiddleware(http.HandlerFunc(a.DeleteNotesHandler))).Methods("DELETE")

	log.Println("Routes established")

}

func (a *App) Run(addr string) {

	// Set up HTTP on Gorilla mux for a graceful shutdown
	srv := &http.Server{
		Addr:         addr,
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
