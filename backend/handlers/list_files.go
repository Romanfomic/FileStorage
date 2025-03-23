package handlers

import (
	"encoding/json"
	"net/http"

	"backend/config"
	"backend/middleware"
)

type FileMetadata struct {
	FileID     int    `json:"file_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	FullPath   string `json:"full_path"`
	CreateDate string `json:"create_date"`
	EditDate   string `json:"edit_date"`
	VersionID  int    `json:"version_id"`
	GroupID    *int   `json:"group_id"`
	OwnerID    *int   `json:"owner_id"`
	AccessID   *int   `json:"access_id"`
	RoleID     *int   `json:"role_id"`
}

func GetUserFiles(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT 
			file_id, name, type, full_path, create_date, edit_date,
			version_id, group_id, owner_id, access_id, role_id
		FROM Files
		WHERE owner_id = $1
	`

	rows, err := config.PostgresDB.Query(query, userID)
	if err != nil {
		http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []FileMetadata
	for rows.Next() {
		var file FileMetadata
		err := rows.Scan(
			&file.FileID,
			&file.Name,
			&file.Type,
			&file.FullPath,
			&file.CreateDate,
			&file.EditDate,
			&file.VersionID,
			&file.GroupID,
			&file.OwnerID,
			&file.AccessID,
			&file.RoleID,
		)
		if err != nil {
			http.Error(w, "Row scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		files = append(files, file)
	}

	if len(files) == 0 {
		http.Error(w, "There are no files", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(files); err != nil {
		http.Error(w, "Encoding error", http.StatusInternalServerError)
		return
	}
}
