package users

import (
	"Trames/internal/core/db"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

func (u *UserHandle) LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(input.Username)
	if username == "" || input.Password == "" {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	dbUser, err := u.GetUserByUsername(ctx, username)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	match, err := dbUser.Password.Matches(input.Password)
	if err != nil {
		u.Logger.Error("Error trying to match passwords: "+err.Error(), map[string]string{
			"username": username,
		})
		http.Error(w, "There was a problem logging you in", http.StatusInternalServerError)
		return
	}
	if !match {
		u.Logger.Info("Password and username combination doesn't exist", map[string]string{
			"username": username,
		})
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	sessionToken, err := GenerateSessionToken()
	if err != nil {
		u.Logger.Error("Error trying to generate session token: "+err.Error(), nil)
		http.Error(w, "There was a problem logging you in", http.StatusInternalServerError)
		return
	}
	expiration := sessionExpiration(u.SessionConfig)
	expiresAt := time.Now().UTC().Add(expiration)
	_, err = u.Queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:    dbUser.ID,
		TokenHash: HashSessionToken(sessionToken),
		ExpiresAt: expiresAt.Format(sqliteTimeLayout),
	})
	if err != nil {
		u.Logger.Error("Error trying to create session: "+err.Error(), nil)
		http.Error(w, "There was a problem logging you in", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		MaxAge:   int(expiration.Seconds()),
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   secureCookie(r),
		SameSite: http.SameSiteLaxMode,
	})
	WriteJSON(w, http.StatusOK, dbUser)
	u.Logger.Info("Logged in successfully", map[string]string{"username": dbUser.Username})
}

func (u *UserHandle) CurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionToken, err := SessionTokenFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized status", http.StatusUnauthorized)
		return
	}

	session, err := u.Queries.FindSessionByTokenHash(ctx, HashSessionToken(sessionToken))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Unauthorized status", http.StatusUnauthorized)
			return
		}
		u.Logger.Error("Error trying to find session: "+err.Error(), nil)
		http.Error(w, "There was a problem and we couldn't fulfill your request", http.StatusInternalServerError)
		return
	}

	dbUser, err := u.GetUserByUserID(ctx, session.UserID)
	if err != nil {
		u.Logger.Error("Error trying to get user: "+err.Error(), nil)
		http.Error(w, "There was a problem and we couldn't fulfill your request", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusOK, dbUser)
	u.Logger.Info("User retrieved successfully", map[string]string{"username": dbUser.Username})
}
