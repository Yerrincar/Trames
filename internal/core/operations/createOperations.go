package operations

import (
	"Trames/internal/core/db"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func (h *Handler) CreateTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	input, err := parseForm(r, "task")
	if err != nil {
		http.Error(w, "Error trying to parse form input", http.StatusBadRequest)
		return

	}

	projectId, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil {
		http.Error(w, "Error trying to parse projectId", http.StatusBadRequest)
		return

	}
	_, err = h.Queries.SelectProjectsByUserAndProjectId(ctx, db.SelectProjectsByUserAndProjectIdParams{
		UserID: userId,
		ID:     int64(projectId),
	})
	if err == sql.ErrNoRows {
		h.Logger.Error("No sub projects with that ID: "+err.Error(), nil)
		http.Error(w, "No sub projects with that ID: ", http.StatusBadRequest)
		return
	}
	if err != nil {
		h.Logger.Error("Error trying to select project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to select project in the database", http.StatusBadRequest)
		return
	}
	var response db.Task

	if input.SubProject != "" {
		subProjectId, err := h.Queries.SelectSubProjectIdBySubProjectName(ctx, db.SelectSubProjectIdBySubProjectNameParams{UserID: userId,
			ProjectID: int64(projectId), SubProject: input.SubProject})
		if err != nil {
			h.Logger.Error("Error trying to select sub projects in the database: "+err.Error(), nil)
			http.Error(w, "Error trying to select sub projects in the database", http.StatusBadRequest)
			return
		}
		response, err = h.Queries.InsertTasksByUserProjectAndSubProject(ctx, db.InsertTasksByUserProjectAndSubProjectParams{
			UserID:       userId,
			ProjectID:    int64(projectId),
			SubProjectID: subProjectId,
			Task:         input.Operation,
			Description:  input.Description,
			Status:       input.Status,
			Priority:     input.Priority,
		})
		if err != nil {
			h.Logger.Error("Error trying to insert task in the database: "+err.Error(), nil)
			http.Error(w, "Error trying to insert task in the database", http.StatusBadRequest)
			return
		}
	} else {
		response, err = h.Queries.InsertTasksByUserAndProject(ctx, db.InsertTasksByUserAndProjectParams{
			UserID:      userId,
			ProjectID:   int64(projectId),
			Task:        input.Operation,
			Description: input.Description,
			Status:      input.Status,
			Priority:    input.Priority,
		})
		if err != nil {
			h.Logger.Error("Error trying to insert task in the database: "+err.Error(), nil)
			http.Error(w, "Error trying to insert task in the database", http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) CreateSubProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	input, err := parseForm(r, "project")
	if err != nil {
		http.Error(w, "Error trying to parse form input", http.StatusBadRequest)
		return

	}

	projectId, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil {
		http.Error(w, "Error trying to convert projectId", http.StatusBadRequest)
		return
	}

	_, err = h.Queries.SelectProjectsByUserAndProjectId(ctx, db.SelectProjectsByUserAndProjectIdParams{
		UserID: userId,
		ID:     int64(projectId),
	})

	if err == sql.ErrNoRows {
		h.Logger.Error("No sub projects with that ID: "+err.Error(), nil)
		http.Error(w, "No sub projects with that ID: ", http.StatusBadRequest)
		return
	}
	if err != nil {
		h.Logger.Error("Error trying to select project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to select project in the database", http.StatusBadRequest)
		return
	}
	response, err := h.Queries.InsertSubProject(ctx, db.InsertSubProjectParams{
		UserID:      userId,
		ProjectID:   int64(projectId),
		SubProject:  input.Operation,
		Description: input.Description,
		Status:      input.Status,
	})
	if err != nil {
		h.Logger.Error("Error trying to insert sub-project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to insert sub-project in the database", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	input, err := parseForm(r, "project")
	if err != nil {
		http.Error(w, "Error trying to parse form input", http.StatusBadRequest)
		return

	}
	response, err := h.Queries.InsertProjectsByUserAndProject(ctx, db.InsertProjectsByUserAndProjectParams{
		UserID:      userId,
		Project:     input.Operation,
		Description: input.Description,
		Status:      input.Status,
	})
	if err != nil {
		h.Logger.Error("Error trying to insert project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to insert project in the database", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	projectId, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil {
		http.Error(w, "Error trying to parse projectId", http.StatusBadRequest)
		return
	}

	input, err := parseForm(r, "project")
	if err != nil {
		http.Error(w, "Error trying to parse form input", http.StatusBadRequest)
		return
	}

	err = h.Queries.UpdateProjectsByUserAndProject(ctx, db.UpdateProjectsByUserAndProjectParams{
		Project:     input.Operation,
		Description: input.Description,
		Status:      input.Status,
		ID:          int64(projectId),
		UserID:      userId,
	})
	if err != nil {
		h.Logger.Error("Error trying to update project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to update project in the database", http.StatusBadRequest)
		return
	}

	response, err := h.Queries.SelectProjectsByUserAndProjectId(ctx, db.SelectProjectsByUserAndProjectIdParams{
		UserID: userId,
		ID:     int64(projectId),
	})
	if err != nil {
		h.Logger.Error("Error trying to select updated project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to select updated project in the database", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) UpdateSubProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	projectId, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil {
		http.Error(w, "Error trying to parse projectId", http.StatusBadRequest)
		return
	}

	subProjectId, err := strconv.Atoi(r.URL.Query().Get("subProjectId"))
	if err != nil {
		http.Error(w, "Error trying to parse subProjectId", http.StatusBadRequest)
		return
	}

	input, err := parseForm(r, "project")
	if err != nil {
		http.Error(w, "Error trying to parse form input", http.StatusBadRequest)
		return
	}

	err = h.Queries.UpdateSubProjectsByUserAndProject(ctx, db.UpdateSubProjectsByUserAndProjectParams{
		SubProject:  input.Operation,
		Description: input.Description,
		Status:      input.Status,
		ID:          int64(subProjectId),
		UserID:      userId,
		ProjectID:   int64(projectId),
	})
	if err != nil {
		h.Logger.Error("Error trying to update sub-project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to update sub-project in the database", http.StatusBadRequest)
		return
	}

	rows, err := h.Queries.SelectSubProjectsByUserAndProject(ctx, db.SelectSubProjectsByUserAndProjectParams{
		UserID:    userId,
		ProjectID: int64(projectId),
	})
	if err != nil {
		h.Logger.Error("Error trying to select updated sub-project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to select updated sub-project in the database", http.StatusBadRequest)
		return
	}

	for _, row := range rows {
		if row.ID == int64(subProjectId) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(row)
			return
		}
	}

	http.Error(w, "No sub projects with that ID", http.StatusBadRequest)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	projectId, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil {
		http.Error(w, "Error trying to parse projectId", http.StatusBadRequest)
		return
	}

	taskId, err := strconv.Atoi(r.URL.Query().Get("taskId"))
	if err != nil {
		http.Error(w, "Error trying to parse taskId", http.StatusBadRequest)
		return
	}

	input, err := parseForm(r, "task")
	if err != nil {
		http.Error(w, "Error trying to parse form input", http.StatusBadRequest)
		return
	}

	rows, err := h.Queries.SelectTasksByUserAndProject(ctx, db.SelectTasksByUserAndProjectParams{
		UserID:    userId,
		ProjectID: int64(projectId),
	})
	if err != nil {
		h.Logger.Error("Error trying to select tasks by user in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to select tasks by user", http.StatusInternalServerError)
		return
	}

	taskFound := false
	for _, row := range rows {
		if row.ID == int64(taskId) {
			taskFound = true
			break
		}
	}
	if !taskFound {
		http.Error(w, "No tasks with that ID", http.StatusBadRequest)
		return
	}

	err = h.Queries.UpdateTasksByUserAndProject(ctx, db.UpdateTasksByUserAndProjectParams{
		Task:        input.Operation,
		Description: input.Description,
		Status:      input.Status,
		Priority:    input.Priority,
		ID:          int64(taskId),
		UserID:      userId,
	})
	if err != nil {
		h.Logger.Error("Error trying to update task in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to update task in the database", http.StatusBadRequest)
		return
	}

	rows, err = h.Queries.SelectTasksByUserAndProject(ctx, db.SelectTasksByUserAndProjectParams{
		UserID:    userId,
		ProjectID: int64(projectId),
	})
	if err != nil {
		h.Logger.Error("Error trying to select updated task in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to select updated task", http.StatusInternalServerError)
		return
	}

	for _, row := range rows {
		if row.ID == int64(taskId) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(row)
			return
		}
	}

	http.Error(w, "No tasks with that ID", http.StatusBadRequest)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	projectId, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil {
		http.Error(w, "Error trying to parse projectId", http.StatusBadRequest)
		return

	}

	taskId, err := strconv.Atoi(r.URL.Query().Get("taskId"))
	if err != nil {
		http.Error(w, "Error trying to parse taskId", http.StatusBadRequest)
		return

	}

	_, err = h.Queries.SelectProjectsByUserAndProjectId(ctx, db.SelectProjectsByUserAndProjectIdParams{
		UserID: userId,
		ID:     int64(projectId),
	})
	if err == sql.ErrNoRows {
		h.Logger.Error("No projects with that ID: "+err.Error(), nil)
		http.Error(w, "No projects with that ID: ", http.StatusBadRequest)
		return
	}
	if err != nil {
		h.Logger.Error("Error trying to select project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to select project in the database", http.StatusBadRequest)
		return
	}

	taskFound := false
	var subProjectID interface{}
	subProjectIdParam := r.URL.Query().Get("subProjectId")
	if subProjectIdParam != "" {
		subProjectId, err := strconv.Atoi(subProjectIdParam)
		if err != nil {
			http.Error(w, "Error trying to parse subProjectId", http.StatusBadRequest)
			return

		}
		subProjectID = int64(subProjectId)

		rows, err := h.Queries.SelectTasksByUserAndProjectAndSubProject(ctx, db.SelectTasksByUserAndProjectAndSubProjectParams{
			UserID:       userId,
			ProjectID:    int64(projectId),
			SubProjectID: int64(subProjectId),
		})
		if err != nil {
			h.Logger.Error("Error trying to select tasks by user in the database: "+err.Error(), nil)
			http.Error(w, "Error trying to select tasks by user", http.StatusInternalServerError)
			return
		}

		for _, row := range rows {
			if row.ID == int64(taskId) {
				taskFound = true
				subProjectID = row.SubProjectID
				break
			}
		}
	} else {
		rows, err := h.Queries.SelectTasksByUserAndProject(ctx, db.SelectTasksByUserAndProjectParams{
			UserID:    userId,
			ProjectID: int64(projectId),
		})
		if err != nil {
			h.Logger.Error("Error trying to select tasks by user in the database: "+err.Error(), nil)
			http.Error(w, "Error trying to select tasks by user", http.StatusInternalServerError)
			return
		}

		for _, row := range rows {
			if row.ID == int64(taskId) {
				taskFound = true
				subProjectID = row.SubProjectID
				break
			}
		}
	}
	if !taskFound {
		h.Logger.Error("No tasks with that ID", nil)
		http.Error(w, "No tasks with that ID", http.StatusBadRequest)
		return
	}

	response, err := h.Queries.DeleteTask(ctx, db.DeleteTaskParams{
		UserID:       userId,
		ID:           int64(taskId),
		SubProjectID: subProjectID,
		ProjectID:    int64(projectId),
	})
	if err != nil {
		h.Logger.Error("Error trying to delete task in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to delete task in the database", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) DeleteSubProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	projectId, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil {
		http.Error(w, "Error trying to parse projectId", http.StatusBadRequest)
		return

	}

	subProjectId, err := strconv.Atoi(r.URL.Query().Get("subProjectId"))
	if err != nil {
		http.Error(w, "Error trying to parse subProjectId", http.StatusBadRequest)
		return

	}

	_, err = h.Queries.SelectProjectsByUserAndProjectId(ctx, db.SelectProjectsByUserAndProjectIdParams{
		UserID: userId,
		ID:     int64(projectId),
	})
	if err == sql.ErrNoRows {
		h.Logger.Error("No projects with that ID: "+err.Error(), nil)
		http.Error(w, "No projects with that ID: ", http.StatusBadRequest)
		return
	}
	if err != nil {
		h.Logger.Error("Error trying to select project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to select project in the database", http.StatusBadRequest)
		return
	}

	rows, err := h.Queries.SelectSubProjectsByUserAndProject(ctx, db.SelectSubProjectsByUserAndProjectParams{
		UserID:    userId,
		ProjectID: int64(projectId),
	})
	if err != nil {
		h.Logger.Error("Error trying to select sub-projects by user in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to select sub-projects by user", http.StatusInternalServerError)
		return
	}

	subProjectFound := false
	for _, row := range rows {
		if row.ID == int64(subProjectId) {
			subProjectFound = true
			break
		}
	}
	if !subProjectFound {
		h.Logger.Error("No sub projects with that ID", nil)
		http.Error(w, "No sub projects with that ID", http.StatusBadRequest)
		return
	}

	response, err := h.Queries.DeleteSubProjects(ctx, db.DeleteSubProjectsParams{
		UserID:    userId,
		ID:        int64(subProjectId),
		ProjectID: int64(projectId),
	})
	if err != nil {
		h.Logger.Error("Error trying to delete sub-project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to delete sub-project in the database", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	projectId, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil {
		http.Error(w, "Error trying to parse projectId", http.StatusBadRequest)
		return

	}

	_, err = h.Queries.SelectProjectsByUserAndProjectId(ctx, db.SelectProjectsByUserAndProjectIdParams{
		UserID: userId,
		ID:     int64(projectId),
	})
	if err == sql.ErrNoRows {
		h.Logger.Error("No projects with that ID: "+err.Error(), nil)
		http.Error(w, "No projects with that ID: ", http.StatusBadRequest)
		return
	}
	if err != nil {
		h.Logger.Error("Error trying to select project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to select project in the database", http.StatusBadRequest)
		return
	}

	response, err := h.Queries.DeleteProjects(ctx, db.DeleteProjectsParams{
		ID:     int64(projectId),
		UserID: userId,
	})
	if err != nil {
		h.Logger.Error("Error trying to delete project in the database: "+err.Error(), nil)
		http.Error(w, "Error trying to delete project in the database", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) DisplayTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	projectId, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil {
		http.Error(w, "Error trying to parse projectId", http.StatusBadRequest)
		return

	}

	var rows []db.SelectTasksByUserAndProjectRow
	rows, err = h.Queries.SelectTasksByUserAndProject(ctx, db.SelectTasksByUserAndProjectParams{UserID: userId,
		ProjectID: int64(projectId)})
	if err != nil {
		http.Error(w, "Error trying to select tasks by user", http.StatusInternalServerError)
		return
	}
	tasks := make([]*OperationForm, 0)

	for _, row := range rows {
		b := &OperationForm{
			ID:          row.ID,
			Operation:   row.Task,
			SubProject:  row.SubProject.String,
			Description: row.Description,
			Status:      row.Status,
			Priority:    row.Priority,
		}
		tasks = append(tasks, b)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func (h *Handler) DisplayProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	var rows []db.Project
	rows, err = h.Queries.SelectProjectsByUserAndProject(ctx, userId)
	if err != nil {
		http.Error(w, "Error trying to select tasks by user", http.StatusInternalServerError)
		return
	}
	projects := make([]*OperationForm, 0)

	for _, row := range rows {
		b := &OperationForm{
			ID:          row.ID,
			Operation:   row.Project,
			Description: row.Description,
			Status:      row.Status,
		}
		projects = append(projects, b)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(projects); err != nil {
		log.Printf("write json response: %v", err)
	}
}
func (h *Handler) DisplaySubProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, status, err := h.SessionId(ctx, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	projectId, err := strconv.Atoi(r.URL.Query().Get("projectId"))
	if err != nil {
		http.Error(w, "Error trying to parse projectId", http.StatusBadRequest)
		return

	}

	var rows []db.SubProject
	rows, err = h.Queries.SelectSubProjectsByUserAndProject(ctx, db.SelectSubProjectsByUserAndProjectParams{UserID: userId,
		ProjectID: int64(projectId)})
	if err != nil {
		http.Error(w, "Error trying to select tasks by user", http.StatusInternalServerError)
		return
	}
	subProjects := make([]*OperationForm, 0)

	for _, row := range rows {
		b := &OperationForm{
			ID:          row.ID,
			Operation:   row.SubProject,
			Description: row.Description,
			Status:      row.Status,
		}
		subProjects = append(subProjects, b)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(subProjects); err != nil {
		log.Printf("write json response: %v", err)
	}
}
