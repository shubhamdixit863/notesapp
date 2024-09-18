package main

import "notesApp/app"

func main() {
	a := app.App{}
	a.Initialize()
	a.Run("")
}
