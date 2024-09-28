package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"notesApp/models"
	"notesApp/utils"
)

func (a *App) CreateNotesHandler(w http.ResponseWriter, r *http.Request) {

	claims, ok := r.Context().Value("user").(*utils.Claims)
	if !ok {
		utils.CheckInternalServerError(err, w)
		return
	}
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
	createdAt := time.Now()
	fmt.Println(createdAt)

	// insert into the db
	_, err = a.db.Exec(context.Background(), `INSERT INTO notes(name, text, status,delegationUser,completion_time,createdAt) VALUES($1, $2, $3,$4,$5,$6)`,
		notes.Name, notes.Text, notes.Status, claims.Username, notes.CompletionTime, createdAt)
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
		err := rows.Scan(&note.Id, &note.Name, &note.Text, &note.Status, &note.CompletionTime, &note.DelegationUser, &note.CreatedAt, &note.SharedUsers) // Adjust fields as per your table schema
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
	err := row.Scan(&note.Id, &note.Name, &note.Text, &note.Status, &note.CompletionTime, &note.DelegationUser, &note.CreatedAt, &note.SharedUsers) // Adjust fields as per your table schema
	if err != nil {
		log.Println("Error scanning note:", err)
		utils.CheckInternalServerError(err, w)
		return
	}
	res := models.NewResponse("200", "Notes fetched successfully", nil, note)
	utils.RespondWithJSON(w, 200, res)

}

func (a *App) UpdateNotesHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*utils.Claims)
	if !ok {
		utils.CheckInternalServerError(err, w)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	row := a.db.QueryRow(context.Background(), `SELECT delegationuser FROM notes where id = $1`, id)
	if err != nil {
		log.Println("Error getting note:", err)
		utils.CheckInternalServerError(err, w)
		return
	}
	var delegateUser string
	err := row.Scan(&delegateUser) // Adjust fields as per your table schema
	if err != nil {
		log.Println("Error scanning note:", err)
		utils.CheckInternalServerError(err, w)
		return
	}

	// check if the user has allowed access to edit the note
	if delegateUser != claims.Username {
		utils.CheckInternalServerError(errors.New("not Allowed To Edit"), w)
		return
	}

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

	// update the data in the db

	_, err = a.db.Exec(context.Background(), `UPDATE notes SET name=$1, text=$2, status=$3, delegationUser=$4, completion_time=$5 WHERE id=$6`,
		notes.Name, notes.Text, notes.Status, claims.Username, notes.CompletionTime, id)
	if err != nil {
		log.Println(err)
		utils.CheckInternalServerError(err, w)
		return
	}
	res := models.NewResponse("200", "Note updated successfully", nil, nil)
	utils.RespondWithJSON(w, 200, res)
}

func (a *App) DeleteNotesHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*utils.Claims)
	if !ok {
		utils.CheckInternalServerError(err, w)
		return
	}
	vars := mux.Vars(r)

	id := vars["id"]

	row := a.db.QueryRow(context.Background(), `SELECT delegationuser FROM notes where id = $1`, id)
	if err != nil {
		log.Println("Error getting note:", err)
		utils.CheckInternalServerError(err, w)
		return
	}
	var delegateUser string
	err := row.Scan(&delegateUser) // Adjust fields as per your table schema
	if err != nil {
		log.Println("Error scanning note:", err)
		utils.CheckInternalServerError(err, w)
		return
	}

	// check if the user has allowed access to edit the note
	if delegateUser != claims.Username {
		utils.CheckInternalServerError(errors.New("not Allowed To delete"), w)
		return
	}

	_, err = a.db.Exec(context.Background(), `DELETE FROM notes WHERE id=$1`, id)
	if err != nil {
		log.Println(err)
		utils.CheckInternalServerError(err, w)
		return
	}

	// Respond with a success message
	res := models.NewResponse("200", "Note deleted successfully", nil, nil)
	utils.RespondWithJSON(w, 200, res)

}

func (a *App) SearchNotesHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*utils.Claims)
	if !ok {
		utils.CheckInternalServerError(err, w)
		return
	}

	// Extract query parameters
	textPattern := r.URL.Query().Get("textPattern")
	name := r.URL.Query().Get("name")
	status := r.URL.Query().Get("status")
	completionTime := r.URL.Query().Get("completionTime")

	query := `SELECT id, name, text, status, delegationUser, completion_time ,sharedusers
			  FROM notes WHERE delegationUser = $1`
	args := []interface{}{claims.Username}

	if textPattern != "" {
		query += " AND text ILIKE '%' || $2 || '%'"
		args = append(args, textPattern)
	}
	if name != "" {
		query += " AND name ILIKE '%' || $3 || '%'"
		args = append(args, name)
	}
	if status != "" {
		query += " AND status = $4"
		args = append(args, status)
	}
	if completionTime != "" {
		query += " AND completion_time = $5"
		args = append(args, completionTime)
	}

	rows, err := a.db.Query(context.Background(), query, args...)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		err := rows.Scan(&note.Id, &note.Name, &note.Text, &note.Status, &note.DelegationUser, &note.CompletionTime, &note.SharedUsers)
		if err != nil {
			log.Println("Error scanning note: hii", err)
			utils.CheckInternalServerError(err, w)
			return
		}
		notes = append(notes, note)
	}

	res := models.NewResponse("200", "Notes retrieved successfully", nil, notes)
	utils.RespondWithJSON(w, 200, res)
}

func (a *App) AnalyzeNoteHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*utils.Claims)
	if !ok {
		utils.CheckInternalServerError(err, w)
		return
	}

	vars := mux.Vars(r)
	noteID := vars["id"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}
	var requestBody struct {
		Pattern string `json:"pattern"`
	}
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	var noteText string
	err = a.db.QueryRow(context.Background(), "SELECT text FROM notes WHERE id = $1 AND delegationUser = $2", noteID, claims.Username).Scan(&noteText)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	matches := regexp.MustCompile(requestBody.Pattern).FindAllString(noteText, -1)

	analysisResult := map[string]interface{}{
		"count":       len(matches),
		"matchedText": matches,
	}

	res := models.NewResponse("200", "Note analyzed successfully", nil, analysisResult)
	utils.RespondWithJSON(w, 200, res)
}

func (a *App) ShareNoteHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*utils.Claims)
	if !ok {
		utils.CheckInternalServerError(err, w)
		return
	}
	vars := mux.Vars(r)
	noteID := vars["id"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	var shareRequest struct {
		SharedUsers []models.SharedUser `json:"sharedUsers"`
	}

	err = json.Unmarshal(body, &shareRequest)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	var owner string
	err = a.db.QueryRow(context.Background(), "SELECT delegationUser FROM notes WHERE id = $1", noteID).Scan(&owner)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	if owner != claims.Username {
		utils.RespondWithError(w, http.StatusForbidden, "You are not the owner of this note")
		return
	}

	sharedUsersJSON, err := json.Marshal(shareRequest.SharedUsers)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	_, err = a.db.Exec(context.Background(), `
		UPDATE notes SET sharedUsers = $1 WHERE id = $2`, sharedUsersJSON, noteID)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	res := models.NewResponse("200", "Note sharing settings updated successfully", nil, nil)
	utils.RespondWithJSON(w, 200, res)
}

func (a *App) GetSharedUsersHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*utils.Claims)
	if !ok {
		utils.CheckInternalServerError(err, w)
		return
	}

	vars := mux.Vars(r)
	noteID := vars["id"]

	var owner string
	err := a.db.QueryRow(context.Background(), "SELECT delegationUser FROM notes WHERE id = $1", noteID).Scan(&owner)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	if owner != claims.Username {
		utils.RespondWithError(w, http.StatusForbidden, "You are not the owner of this note")
		return
	}

	var sharedUsersJSON string
	err = a.db.QueryRow(context.Background(), "SELECT sharedUsers FROM notes WHERE id = $1", noteID).Scan(&sharedUsersJSON)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	var sharedUsers []models.SharedUser
	err = json.Unmarshal([]byte(sharedUsersJSON), &sharedUsers)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	res := models.NewResponse("200", "Shared users retrieved successfully", nil, sharedUsers)
	utils.RespondWithJSON(w, 200, res)
}

func (a *App) UpdateTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*utils.Claims)
	if !ok {
		utils.CheckInternalServerError(err, w)
		return
	}

	vars := mux.Vars(r)
	noteID := vars["id"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	var statusUpdate struct {
		Status         string `json:"status"`
		DelegationUser string `json:"delegationUser,omitempty"`
	}

	err = json.Unmarshal(body, &statusUpdate)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	var owner string
	err = a.db.QueryRow(context.Background(), "SELECT delegationUser FROM notes WHERE id = $1", noteID).Scan(&owner)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	if owner != claims.Username {
		utils.RespondWithError(w, http.StatusForbidden, "You are not the owner of this note")
		return
	}

	query := `UPDATE notes SET status = $1`
	args := []interface{}{statusUpdate.Status}

	if statusUpdate.DelegationUser != "" {
		query += `, delegationUser = $2 WHERE id = $3`
		args = append(args, statusUpdate.DelegationUser, noteID)
	} else {
		query += ` WHERE id = $2`
		args = append(args, noteID)
	}

	_, err = a.db.Exec(context.Background(), query, args...)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	res := models.NewResponse("200", "Task status updated successfully", nil, nil)
	utils.RespondWithJSON(w, 200, res)
}
