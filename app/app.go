package app

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" //use pgx in database/sql mode
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	err  error
	wait time.Duration
)

type App struct {
	Router   *mux.Router
	db       *pgxpool.Pool
	bindport string
	username string
	role     string
}

func NewApp() *App {
	return &App{}
}

func (a *App) Initialize() {

	connStr := os.Getenv("CONN_STRING")
	db, err := pgxpool.New(context.Background(), connStr)
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
	a.Router.HandleFunc("/users", a.GetAllUsersHandler).Methods("GET")
	a.Router.Handle("/notes", JWTAuthMiddleware(http.HandlerFunc(a.CreateNotesHandler))).Methods("POST")
	a.Router.Handle("/notes", JWTAuthMiddleware(http.HandlerFunc(a.GetNotesHandler))).Methods("GET")
	a.Router.Handle("/notes/search", JWTAuthMiddleware(http.HandlerFunc(a.SearchNotesHandler))).Methods("GET")
	a.Router.Handle("/notes/{id}/analyze", JWTAuthMiddleware(http.HandlerFunc(a.AnalyzeNoteHandler))).Methods("POST")
	a.Router.Handle("/notes/{id}/share", JWTAuthMiddleware(http.HandlerFunc(a.ShareNoteHandler))).Methods("PUT")
	a.Router.Handle("/notes/{id}/shared-users", JWTAuthMiddleware(http.HandlerFunc(a.GetSharedUsersHandler))).Methods("GET")
	a.Router.Handle("/notes/{id}", JWTAuthMiddleware(http.HandlerFunc(a.UpdateNotesHandler))).Methods("PUT")
	a.Router.Handle("/notes/{id}", JWTAuthMiddleware(http.HandlerFunc(a.DeleteNotesHandler))).Methods("DELETE")
	a.Router.Handle("/notes/{id}", JWTAuthMiddleware(http.HandlerFunc(a.GetNoteById))).Methods("GET")
	a.Router.Handle("/notes/{id}/status", JWTAuthMiddleware(http.HandlerFunc(a.UpdateTaskStatusHandler))).Methods("PUT")

	log.Println("Routes established")

}

func (a *App) Run(addr string) {

	corsOptions := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true,
	})

	// Wrap the router with CORS middleware
	handler := corsOptions.Handler(a.Router)

	// Set up HTTP on Gorilla mux for a graceful shutdown
	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handler, // Assuming a.Router is already set up with the routes
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
		a.db.Close()
	}

	log.Println("Shutting down the application")
	os.Exit(0)
}
