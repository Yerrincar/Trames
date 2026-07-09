package operations

import (
	"Trames/internal/core/users"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
)

func parseForm(r *http.Request, operationType string) (*OperationForm, error) {
	operation := r.FormValue(operationType)
	if strings.TrimSpace(operation) == "" {
		return nil, errors.New("Name is required")
	}

	description := sql.NullString{
		String: r.FormValue("Description"),
		Valid:  true,
	}
	status := r.FormValue("Status")
	if strings.TrimSpace(status) == "" {
		return nil, errors.New("Status is required")
	}

	priority := r.FormValue("Priority")
	if priority == "" {
		priority = "LOW"
	}
	if operationType == "task" {
		subProject := r.FormValue("Sub Project")
		return &OperationForm{
			Operation:   operation,
			SubProject:  subProject,
			Description: description,
			Status:      status,
			Priority:    priority,
		}, nil
	} else {
		return &OperationForm{
			Operation:   operation,
			Description: description,
			Status:      status,
			Priority:    priority,
		}, nil
	}
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
