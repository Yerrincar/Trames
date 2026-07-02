package users

import (
	"Trames/internal/core/db"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	sessionCookieName = "sessionId"
	sqliteTimeLayout  = "2006-01-02 15:04:05"
)

func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func SessionTokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	if cookie.Value == "" {
		return "", errors.New("empty session cookie")
	}
	return cookie.Value, nil
}

func (u *UserHandle) GetUserByUserID(ctx context.Context, id int64) (*UserProfile, error) {
	dbUser, err := u.Queries.SelectUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			u.Logger.Error("The user was not found: "+err.Error(), map[string]string{
				"id": strconv.FormatInt(id, 10),
			})
			return nil, err
		}
		u.Logger.Error("Error trying to select user: "+err.Error(), map[string]string{
			"id": strconv.FormatInt(id, 10),
		})
		return nil, err
	}
	return userProfileFromDB(dbUser), nil
}

func (u *UserHandle) GetUserByUsername(ctx context.Context, username string) (*UserProfile, error) {
	dbUser, err := u.Queries.SelectUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			u.Logger.Error("The username was not found: "+err.Error(), map[string]string{
				"username": username,
			})
			return nil, err
		}
		u.Logger.Error("Error trying to select username: "+err.Error(), map[string]string{
			"username": username,
		})
		return nil, err
	}
	return userProfileFromDB(dbUser), nil
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func userProfileFromDB(dbUser db.User) *UserProfile {
	return &UserProfile{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Password:  password{Hash: dbUser.PasswordHash},
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
}

func validateRegister(username, password string) map[string]string {
	errs := make(map[string]string)
	if username == "" {
		errs["username"] = "username must be provided"
	}
	if password == "" {
		errs["password"] = "password must be provided"
	} else if len(password) < 8 {
		errs["password"] = "password must be at least 8 characters"
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func isUniqueConstraint(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint") || strings.Contains(message, "constraint failed")
}

func sessionExpiration(config SessionConfig) time.Duration {
	if config.SessionExpiration <= 0 {
		return 24 * time.Hour
	}
	return config.SessionExpiration
}

func secureCookie(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}
