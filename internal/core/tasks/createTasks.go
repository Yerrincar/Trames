package tasks

import (
	"Trames/internal/core/db"
	"Trames/internal/core/users"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) CreateTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	input, err := parseForm(r)
	if err != nil {
		http.Error(w, "Error trying to parse form input", http.StatusBadRequest)
		return

	}

	projectId, err := strconv.Atoi(r.URL.Query().Get("projectId"))

	response, err := h.Queries.InsertTasksByUserAndProject(ctx, db.InsertTasksByUserAndProjectParams{
		UserID:      userId,
		ProjectID:   int64(projectId),
		Task:        input.Task,
		Description: input.Description,
		Status:      input.Status,
		Priority:    input.Priority,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func parseForm(r *http.Request) (*TaskForm, error) {
	task := r.FormValue("Task")
	if strings.TrimSpace(task) == "" {
		return nil, errors.New("Task name is required")
	}

	description := sql.NullString{
		String: r.FormValue("Description"),
		Valid:  true,
	}
	status := r.FormValue("Status")

	priority := r.FormValue("Priority")

	return &TaskForm{
		Task:        task,
		Description: description,
		Status:      status,
		Priority:    priority,
	}, nil
}

func (h *Handler) SessionId(ctx context.Context, r *http.Request) (int64, int, error) {
	sessionToken, err := users.SessionTokenFromRequest(r)
	if err != nil {
		return 0, http.StatusUnauthorized, err
	}
	session, err := h.Queries.FindSessionByTokenHash(ctx, users.HashSessionToken(sessionToken))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, http.StatusUnauthorized, err
		}
		h.Logger.Error("Error trying to find session: "+err.Error(), nil)
		return 0, http.StatusInternalServerError, err
	}

	return session.UserID, http.StatusOK, nil
}
