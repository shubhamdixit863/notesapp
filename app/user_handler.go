package app

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"

	"notesApp/models"
	"notesApp/utils"
)

func (a *App) registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "tmpl/register.html")
		return
	}

	// grab user info
	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")

	// no range, bounds, context, type checking
	// Check existence of user
	var user models.User
	err := a.db.QueryRow("SELECT username, password, role FROM users WHERE username=$1",
		username).Scan(&user.Username, &user.Password, &user.Role)
	switch {
	// user is available
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		utils.CheckInternalServerError(err, w)
		// insert to database
		_, err = a.db.Exec(`INSERT INTO users(username, password, role) VALUES($1, $2, $3)`,
			username, hashedPassword, role)
		utils.CheckInternalServerError(err, w)
	case err != nil:
		http.Error(w, "loi: "+err.Error(), http.StatusBadRequest)
		return
	default:
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
}

func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Method %s", r.Method)
	if r.Method != "POST" {
		http.ServeFile(w, r, "tmpl/login.html")
		return
	}

	// grab user info from the submitted form
	username := r.FormValue("usrname")
	password := r.FormValue("psw")

	// query database to get match username
	var user models.User
	err := a.db.QueryRow("SELECT id, username, password FROM users WHERE username=$1",
		username).Scan(&user.Id, &user.Username, &user.Password)
	utils.CheckInternalServerError(err, w)

	// validate password
	/*
		//simple unencrypted method
		if user.Password != password {
			http.Redirect(w, r, "/login", 301)
			return
		}
	*/

	//password is encrypted
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login", 301)
		return
	}

	// Successful login. Create JWT and send it

}
