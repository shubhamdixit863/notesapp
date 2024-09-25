package app

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"

	"notesApp/models"
	"notesApp/utils"
)

func (a *App) CreateNotesHandler(w http.ResponseWriter, r *http.Request) {

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
	_, err = a.db.Exec(context.Background(), `INSERT INTO notes(name, text, status,delegationUser,completion_time) VALUES($1, $2, $3,$4,$5)`,
		notes.Name, notes.Text, notes.Status, notes.DelegationUser, notes.CompletionTime)
	if err != nil {
		log.Println(err)
		utils.CheckInternalServerError(err, w)
		return
	}

	res := models.NewResponse("200", "Notes created successfully", nil, nil)

	utils.RespondWithJSON(w, 200, res)

}

func (a *App) GetNotesHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	rows, err := a.db.Query(context.Background(), `SELECT * FROM notes`)
	if err != nil {
		log.Println("Error querying notes:", err)
		utils.CheckInternalServerError(err, w)
		return
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		err := rows.Scan(&note.Id, &note.Name, &note.Text, &note.Status, &note.CompletionTime, &note.DelegationUser) // Adjust fields as per your table schema
		if err != nil {
			log.Println("Error scanning note:", err)
			utils.CheckInternalServerError(err, w)
			return
		}
		notes = append(notes, note)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating over notes:", err)
		utils.CheckInternalServerError(err, w)
		return
	}

	res := models.NewResponse("200", "Notes fetched successfully", nil, notes)
	utils.RespondWithJSON(w, 200, res)

}
func (a *App) GetNoteById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	row := a.db.QueryRow(context.Background(), `SELECT * FROM notes where id = $1`, id)
	if err != nil {
		log.Println("Error getting note:", err)
		utils.CheckInternalServerError(err, w)
		return
	}
	var note models.Note
	err := row.Scan(&note.Id, &note.Name, &note.Text, &note.Status, &note.CompletionTime, &note.DelegationUser) // Adjust fields as per your table schema
	if err != nil {
		log.Println("Error scanning note:", err)
		utils.CheckInternalServerError(err, w)
		return
	}
	res := models.NewResponse("200", "Notes fetched successfully", nil, note)
	utils.RespondWithJSON(w, 200, res)

}

func (a *App) DeleteNotesHandler(w http.ResponseWriter, r *http.Request) {}
