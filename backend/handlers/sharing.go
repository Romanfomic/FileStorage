package handlers

import (
	"backend/config"
	"backend/middleware"
	"backend/models"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func ShareFileWithUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["file_id"]

	var req struct {
		UserID   int `json:"user_id"`
		AccessID int `json:"access_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := config.PostgresDB.Exec(
		"INSERT INTO File_Users (file_id, user_id, access_id) VALUES ($1, $2, $3) ON CONFLICT (file_id, user_id) DO UPDATE SET access_id = $3",
		fileID, req.UserID, req.AccessID,
	)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func ShareFileWithGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["file_id"]

	var req struct {
		GroupID  int `json:"group_id"`
		AccessID int `json:"access_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := config.PostgresDB.Exec(
		"INSERT INTO File_Groups (file_id, group_id, access_id) VALUES ($1, $2, $3) ON CONFLICT (file_id, group_id) DO UPDATE SET access_id = $3",
		fileID, req.GroupID, req.AccessID,
	)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func GetFilePermissions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["file_id"]

	// get users with access
	userRows, err := config.PostgresDB.Query("SELECT user_id, access_id FROM File_Users WHERE file_id = $1", fileID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer userRows.Close()

	var users []models.FilePermissionUser
	for userRows.Next() {
		var user models.FilePermissionUser
		if err := userRows.Scan(&user.UserID, &user.AccessID); err != nil {
			http.Error(w, "Error scanning users", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	// get groups with access
	groupRows, err := config.PostgresDB.Query("SELECT group_id, access_id FROM File_Groups WHERE file_id = $1", fileID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer groupRows.Close()

	var groups []models.FilePermissionGroup
	for groupRows.Next() {
		var group models.FilePermissionGroup
		if err := groupRows.Scan(&group.GroupID, &group.AccessID); err != nil {
			http.Error(w, "Error scanning groups", http.StatusInternalServerError)
			return
		}
		groups = append(groups, group)
	}

	response := map[string]interface{}{
		"users":  users,
		"groups": groups,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetSharedFiles(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// get files shared to user
	userFilesQuery := `
		SELECT f.file_id, f.name, f.full_path, f.owner_id, fu.access_id, f.version_id, f.create_date, f.edit_date
		FROM Files f
		JOIN File_Users fu ON f.file_id = fu.file_id
		WHERE fu.user_id = $1
	`

	// get files shared to groups
	groupFilesQuery := `
		SELECT f.file_id, f.name, f.full_path, f.owner_id, fg.group_id, fg.access_id, f.version_id, f.create_date, f.edit_date
		FROM Files f
		JOIN File_Groups fg ON f.file_id = fg.file_id
		JOIN Users u ON u.group_id = fg.group_id
		WHERE u.user_id = $1
	`

	// map for merge files data
	filesMap := make(map[int]*models.SharedFile)

	userRows, err := config.PostgresDB.Query(userFilesQuery, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer userRows.Close()

	for userRows.Next() {
		var fileID, ownerID, accessID, versionID int
		var name, fullPath, createDate, editDate string

		err := userRows.Scan(&fileID, &name, &fullPath, &ownerID, &accessID, &versionID, &createDate, &editDate)
		if err != nil {
			http.Error(w, "Error scanning user files", http.StatusInternalServerError)
			return
		}

		filesMap[fileID] = &models.SharedFile{
			FileID:     fileID,
			Name:       name,
			FullPath:   fullPath,
			OwnerID:    ownerID,
			GroupIDs:   nil,
			AccessID:   accessID,
			VersionID:  versionID,
			CreateDate: createDate,
			EditDate:   editDate,
		}
	}

	groupRows, err := config.PostgresDB.Query(groupFilesQuery, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer groupRows.Close()

	for groupRows.Next() {
		var fileID, ownerID, groupID, accessID, versionID int
		var name, fullPath, createDate, editDate string

		err := groupRows.Scan(&fileID, &name, &fullPath, &ownerID, &groupID, &accessID, &versionID, &createDate, &editDate)
		if err != nil {
			http.Error(w, "Error scanning group files", http.StatusInternalServerError)
			return
		}

		// check files exist in map
		if file, exists := filesMap[fileID]; exists {
			// add group_id
			file.GroupIDs = append(file.GroupIDs, groupID)

			// if the group_id was nil, and there was no provision directly to the user, we set it
			if file.GroupIDs == nil {
				file.GroupIDs = []int{groupID}
			}

			// If access is through a higher group, we update the access_id
			if accessID > file.AccessID {
				file.AccessID = accessID
			}
		} else {
			filesMap[fileID] = &models.SharedFile{
				FileID:     fileID,
				Name:       name,
				FullPath:   fullPath,
				OwnerID:    ownerID,
				GroupIDs:   []int{groupID},
				AccessID:   accessID,
				VersionID:  versionID,
				CreateDate: createDate,
				EditDate:   editDate,
			}
		}
	}

	var files []models.SharedFile
	for _, file := range filesMap {
		files = append(files, *file)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func RevokeUserAccess(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["file_id"]
	userID := vars["user_id"]

	_, err := config.PostgresDB.Exec("DELETE FROM File_Users WHERE file_id = $1 AND user_id = $2", fileID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func RevokeGroupAccess(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["file_id"]
	groupID := vars["group_id"]

	_, err := config.PostgresDB.Exec("DELETE FROM File_Groups WHERE file_id = $1 AND group_id = $2", fileID, groupID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
