package main

import "notesApp/app"

func main() {
	a := app.NewApp()
	a.Initialize()
	a.Run("0.0.0.0:8080")
}
