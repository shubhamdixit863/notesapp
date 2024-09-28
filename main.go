package main

import (
	"log"

	"github.com/joho/godotenv"

	"notesApp/app"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	a := app.NewApp()
	a.Initialize()
	a.Run("0.0.0.0:8080")
}
