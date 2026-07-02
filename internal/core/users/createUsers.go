package users

import (
	"Trames/internal/core/db"
	"encoding/json"
	"net/http"
	"strings"
)

func (u *UserHandle) Register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(input.Username)
	if errs := validateRegister(username, input.Password); errs != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]map[string]string{"errors": errs})
		return
	}

	var userPassword password
	if err := userPassword.Set(input.Password); err != nil {
		u.Logger.Error("Error hashing password: "+err.Error(), nil)
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	dbUser, err := u.Queries.InsertUser(r.Context(), db.InsertUserParams{
		Username:     username,
		PasswordHash: userPassword.Hash,
	})
	if err != nil {
		if isUniqueConstraint(err) {
			http.Error(w, "Username already in use", http.StatusConflict)
			return
		}
		u.Logger.Error("Error trying to register user: "+err.Error(), nil)
		http.Error(w, "Error trying to register user", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusCreated, userProfileFromDB(dbUser))
}
