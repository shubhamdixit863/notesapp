package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"notesApp/models"
	"notesApp/utils"
)

func (a *App) CreateNotesHandler(w http.ResponseWriter, r *http.Request) {

	//Get request body

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	var notes models.Note
	// parse Notes body to json

	err = json.Unmarshal(body, &notes)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return

	}

	// insert into the db
	_, err = a.db.Exec(context.Background(), `INSERT INTO notes(name, text, status,delegationUser,completionTime) VALUES($1, $2, $3,$4,$5)`,
		notes.Name, notes.Text, notes.Status, notes.DelegationUser, notes.CompletionTime)
	utils.CheckInternalServerError(err, w)
	utils.RespondWithJSON(w, 200, "success")

}
