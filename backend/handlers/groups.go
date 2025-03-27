package handlers

import (
	"encoding/json"
	"net/http"

	"backend/config"

	"github.com/gorilla/mux"
)

type Group struct {
	ID          int    `json:"group_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	var group Group
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO Groups (name, description) VALUES ($1, $2) RETURNING group_id`
	err := config.PostgresDB.QueryRow(query, group.Name, group.Description).Scan(&group.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

func GetGroups(w http.ResponseWriter, r *http.Request) {
	rows, err := config.PostgresDB.Query("SELECT group_id, name, description FROM Groups")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var group Group
		if err := rows.Scan(&group.ID, &group.Name, &group.Description); err != nil {
			http.Error(w, "Row scan error", http.StatusInternalServerError)
			return
		}
		groups = append(groups, group)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	var group Group
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err := config.PostgresDB.Exec(
		"UPDATE Groups SET name = $1, description = $2 WHERE group_id = $3",
		group.Name, group.Description, groupID,
	)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Проверяем, есть ли файлы, привязанные к группе
	var count int
	err := config.PostgresDB.QueryRow("SELECT COUNT(*) FROM Files WHERE group_id = $1", groupID).Scan(&count)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "Cannot delete group with attached files", http.StatusBadRequest)
		return
	}

	// Удаляем группу
	_, err = config.PostgresDB.Exec("DELETE FROM Groups WHERE group_id = $1", groupID)
	if err != nil {
		http.Error(w, "Failed to delete group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
