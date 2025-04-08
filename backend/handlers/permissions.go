package handlers

import (
	"backend/config"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Permission struct {
	ID          int    `json:"permission_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func GetPermissions(w http.ResponseWriter, r *http.Request) {
	rows, err := config.PostgresDB.Query(`SELECT permission_id, name, description FROM Permissions`)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			http.Error(w, "Row scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		permissions = append(permissions, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}

func GetPermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	permissionID := vars["id"]

	var p Permission
	err := config.PostgresDB.QueryRow(`
		SELECT permission_id, name, description
		FROM Permissions
		WHERE permission_id = $1
	`, permissionID).Scan(&p.ID, &p.Name, &p.Description)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Permission not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}
