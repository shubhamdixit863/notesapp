package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"

	"notesApp/models"
	"notesApp/utils"
)

func (a *App) registerHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	var user models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return

	}
	// no range, bounds, context, type checking
	// Check existence of user
	var existingUser models.User
	err = a.db.QueryRow(context.Background(), "SELECT username, password, role FROM users WHERE username=$1",
		user.Username).Scan(&existingUser.Username, &existingUser.Password, &existingUser.Role)
	switch {
	// user is available
	case errors.Is(err, sql.ErrNoRows):
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		utils.CheckInternalServerError(err, w)
		// insert to database
		_, err = a.db.Exec(context.Background(), `INSERT INTO users(name,username, password, role) VALUES($1, $2, $3,$4)`,
			user.Name, user.Username, hashedPassword, user.Role)
		if err != nil {
			utils.CheckInternalServerError(err, w)
			return
		}
		res := models.NewResponse("200", "User created successfully", nil, nil)
		utils.RespondWithJSON(w, 200, res)
	case err != nil:
		http.Error(w, "loi: "+err.Error(), http.StatusBadRequest)
		return

	}
}

func (a *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Message string `json:"message"`
	}{
		Message: "ok",
	}
	utils.RespondWithJSON(w, 200, payload)
}

func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	// Read the body of the request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	// Parse request body to get user credentials
	var reqUser models.User
	err = json.Unmarshal(body, &reqUser)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	// Query database to find the user by username
	var dbUser models.User
	err = a.db.QueryRow(context.Background(), "SELECT id, username, password, role FROM users WHERE username=$1", reqUser.Username).Scan(&dbUser.Id, &dbUser.Username, &dbUser.Password, &dbUser.Role)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		res := models.NewResponse("401", "Invalid username or password", nil, nil)
		utils.RespondWithJSON(w, http.StatusUnauthorized, res)
		return
	case err != nil:
		utils.CheckInternalServerError(err, w)
		return
	}

	// Compare the provided password with the hashed password from the database
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(reqUser.Password))
	if err != nil {
		// Password mismatch
		res := models.NewResponse("401", "Invalid username or password", nil, nil)
		utils.RespondWithJSON(w, http.StatusUnauthorized, res)
		return
	}

	// Successful login: Generate JWT (this part depends on your JWT setup)
	token, err := utils.GenerateJWT(dbUser.Username, dbUser.Role)
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	// Send success response with token
	res := models.NewResponse("200", "Login successful", nil, map[string]string{"token": token})
	utils.RespondWithJSON(w, http.StatusOK, res)
}
func (a *App) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	rows, err := a.db.Query(context.Background(), "SELECT name,id, username, role FROM users")
	if err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.Name, &user.Id, &user.Username, &user.Role) // Exclude password for security
		if err != nil {
			utils.CheckInternalServerError(err, w)
			return
		}
		users = append(users, user)
	}

	// Check for errors during rows iteration
	if err := rows.Err(); err != nil {
		utils.CheckInternalServerError(err, w)
		return
	}

	// Send the response with the list of users
	res := models.NewResponse("200", "Users retrieved successfully", users, nil)
	utils.RespondWithJSON(w, http.StatusOK, res)
}
