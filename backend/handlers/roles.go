package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"backend/config"

	"github.com/gorilla/mux"
)

type Role struct {
	ID          int    `json:"role_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Permissions []int  `json:"permissions"`
}

func CreateRole(w http.ResponseWriter, r *http.Request) {
	var role Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO Roles (name, description) VALUES ($1, $2) RETURNING role_id`
	err := config.PostgresDB.QueryRow(query, role.Name, role.Description).Scan(&role.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(role)
}

func GetRoles(w http.ResponseWriter, r *http.Request) {
	rows, err := config.PostgresDB.Query("SELECT role_id, name, description FROM Roles")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			http.Error(w, "Row scan error", http.StatusInternalServerError)
			return
		}

		// get permissions
		permRows, err := config.PostgresDB.Query(`
			SELECT p.permission_id
			FROM Role_Permissions rp
			JOIN Permissions p ON rp.permission_id = p.permission_id
			WHERE rp.role_id = $1
		`, role.ID)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer permRows.Close()

		var permissions []int
		for permRows.Next() {
			var permID int
			if err := permRows.Scan(&permID); err != nil {
				http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			permissions = append(permissions, permID)
		}
		role.Permissions = permissions

		roles = append(roles, role)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

func GetRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	// get role
	var role Role
	query := `
		SELECT role_id, name, description
		FROM Roles
		WHERE role_id = $1
	`
	err := config.PostgresDB.QueryRow(query, roleID).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Role not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// get permissions
	rows, err := config.PostgresDB.Query(`
		SELECT p.permission_id
		FROM Role_Permissions rp
		JOIN Permissions p ON rp.permission_id = p.permission_id
		WHERE rp.role_id = $1
	`, roleID)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var permissions []int
	for rows.Next() {
		var permID int
		if err := rows.Scan(&permID); err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		permissions = append(permissions, permID)
	}

	role.Permissions = permissions

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(role); err != nil {
		http.Error(w, "Encoding error", http.StatusInternalServerError)
	}
}

func UpdateRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	var role Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// update name and description
	_, err := config.PostgresDB.Exec(
		"UPDATE Roles SET name = $1, description = $2 WHERE role_id = $3",
		role.Name, role.Description, roleID,
	)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// delete old permissions
	_, err = config.PostgresDB.Exec("DELETE FROM Role_Permissions WHERE role_id = $1", roleID)
	if err != nil {
		http.Error(w, "Failed to update permissions", http.StatusInternalServerError)
		return
	}

	// add new permissions
	for _, permID := range role.Permissions {
		_, err = config.PostgresDB.Exec(
			"INSERT INTO Role_Permissions (role_id, permission_id) VALUES ($1, $2)",
			roleID, permID,
		)
		if err != nil {
			http.Error(w, "Failed to insert permissions", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.URL.Query().Get("id")
	query := "DELETE FROM Roles WHERE role_id = $1"
	_, err := config.PostgresDB.Exec(query, roleID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
