package handlers

import (
	"backend/config"
	"backend/middleware"
	"backend/models"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func generateUniqueVersionName(baseName string, fileID int) (string, error) {
	name := baseName
	count := 1

	for {
		var existingID int
		err := config.PostgresDB.QueryRow(`
            SELECT version_id FROM FileVersions 
            WHERE file_id = $1 AND name = $2
        `, fileID, name).Scan(&existingID)

		if err == sql.ErrNoRows {
			return name, nil
		} else if err != nil {
			return "", err
		}

		count++
		name = fmt.Sprintf("%s_%d", baseName, count)
	}
}

func CreateFileVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileIDStr := vars["file_id"]
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusUnauthorized)
		return
	}

	type FileVersionRequest struct {
		Name string `json:"name"`
	}

	var requestBody FileVersionRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	requestedName := requestBody.Name
	if requestedName == "" {
		requestedName = "version"
	}

	// check if user is owner
	var ownerID int
	err = config.PostgresDB.QueryRow(`
        SELECT owner_id FROM Files WHERE file_id = $1
    `, fileID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if ownerID != userID {
		http.Error(w, "You are not the owner of the file", http.StatusForbidden)
		return
	}

	// get current mongo_file_id
	var currentMongoFileID string
	err = config.PostgresDB.QueryRow(`
        SELECT mongo_file_id FROM Files WHERE file_id = $1
    `, fileID).Scan(&currentMongoFileID)
	if err != nil {
		http.Error(w, "Failed to get current file data", http.StatusInternalServerError)
		return
	}

	// create new mongo copy
	bucket, err := gridfs.NewBucket(config.DB)
	if err != nil {
		http.Error(w, "GridFS initialization error", http.StatusInternalServerError)
		return
	}

	sourceID, err := primitive.ObjectIDFromHex(currentMongoFileID)
	if err != nil {
		http.Error(w, "Invalid MongoDB file ID", http.StatusInternalServerError)
		return
	}

	downloadStream, err := bucket.OpenDownloadStream(sourceID)
	if err != nil {
		http.Error(w, "Failed to open source file", http.StatusInternalServerError)
		return
	}
	defer downloadStream.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, downloadStream); err != nil {
		http.Error(w, "Failed to read source file", http.StatusInternalServerError)
		return
	}

	uploadStream, err := bucket.OpenUploadStream(
		requestedName,
		options.GridFSUpload().SetMetadata(bson.M{"owner_id": userID}),
	)
	if err != nil {
		http.Error(w, "Failed to create new file in MongoDB", http.StatusInternalServerError)
		return
	}
	defer uploadStream.Close()

	if _, err := io.Copy(uploadStream, bytes.NewReader(buf.Bytes())); err != nil {
		http.Error(w, "Failed to upload new file", http.StatusInternalServerError)
		return
	}

	newMongoID := uploadStream.FileID.(primitive.ObjectID)
	newMongoFileIDStr := newMongoID.Hex()

	// generate new uniq version name
	uniqueName, err := generateUniqueVersionName(requestedName, fileID)
	if err != nil {
		http.Error(w, "Error generating unique version name", http.StatusInternalServerError)
		return
	}

	// create new version
	var versionID int
	err = config.PostgresDB.QueryRow(`
        INSERT INTO FileVersions (file_id, user_id, name, mongo_file_id)
        VALUES ($1, $2, $3, $4)
        RETURNING version_id
    `, fileID, userID, uniqueName, newMongoFileIDStr).Scan(&versionID)
	if err != nil {
		http.Error(w, "Failed to create file version", http.StatusInternalServerError)
		return
	}

	// update curent file version
	_, err = config.PostgresDB.Exec(`
        UPDATE Files SET version_id = $1, mongo_file_id = $2, edit_date = CURRENT_TIMESTAMP
        WHERE file_id = $3
    `, versionID, newMongoFileIDStr, fileID)
	if err != nil {
		http.Error(w, "Failed to update file with new version", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "New version created",
		"version_id": versionID,
	})
}

func GetFileVersions(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	fileID, err := strconv.Atoi(vars["file_id"])
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	// сheck if user is owner
	var ownerID int
	err = config.PostgresDB.QueryRow(`SELECT owner_id FROM Files WHERE file_id = $1`, fileID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if ownerID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// get versions
	rows, err := config.PostgresDB.Query(`
        SELECT version_id, name, create_date, edit_date
        FROM FileVersions
        WHERE file_id = $1
        ORDER BY create_date DESC
    `, fileID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var versions []models.FileVersion
	for rows.Next() {
		var v models.FileVersion
		err := rows.Scan(&v.VersionID, &v.Name, &v.CreateDate, &v.EditDate)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		versions = append(versions, v)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

func UpdateFileVersionName(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	versionID, err := strconv.Atoi(vars["version_id"])
	if err != nil {
		http.Error(w, "Invalid version ID", http.StatusBadRequest)
		return
	}

	// check if user owns the version
	var dbUserID int
	err = config.PostgresDB.QueryRow(`
		SELECT user_id FROM FileVersions WHERE version_id = $1
	`, versionID).Scan(&dbUserID)
	if err != nil {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}
	if dbUserID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// read new name from body
	var req struct {
		NewName string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.NewName == "" {
		http.Error(w, "New name cannot be empty", http.StatusBadRequest)
		return
	}

	// update name
	_, err = config.PostgresDB.Exec(`
		UPDATE FileVersions
		SET name = $1, edit_date = NOW()
		WHERE version_id = $2
	`, req.NewName, versionID)
	if err != nil {
		http.Error(w, "Database update error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Version name updated successfully",
	})
}

func DeleteFileVersion(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	versionID, err := strconv.Atoi(vars["version_id"])
	if err != nil {
		http.Error(w, "Invalid version ID", http.StatusBadRequest)
		return
	}

	// get version
	var mongoFileID string
	var dbUserID int
	var fileID int
	err = config.PostgresDB.QueryRow(`
		SELECT user_id, mongo_file_id, file_id
		FROM FileVersions
		WHERE version_id = $1
	`, versionID).Scan(&dbUserID, &mongoFileID, &fileID)
	if err != nil {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}
	if dbUserID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// check if version is not current
	var currentVersionID int
	err = config.PostgresDB.QueryRow(`
		SELECT version_id
		FROM Files
		WHERE file_id = $1
	`, fileID).Scan(&currentVersionID)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if currentVersionID == versionID {
		http.Error(w, "Cannot delete current active version", http.StatusForbidden)
		return
	}

	bucket, err := gridfs.NewBucket(config.DB)
	if err != nil {
		http.Error(w, "GridFS initialization error", http.StatusInternalServerError)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(mongoFileID)
	if err != nil {
		http.Error(w, "Invalid mongo file ID", http.StatusInternalServerError)
		return
	}

	// deelte mongo file
	err = bucket.Delete(objectID)
	if err != nil {
		http.Error(w, "Failed to delete file from storage", http.StatusInternalServerError)
		return
	}

	// delete version
	_, err = config.PostgresDB.Exec(`
		DELETE FROM FileVersions
		WHERE version_id = $1
	`, versionID)
	if err != nil {
		http.Error(w, "Failed to delete version from database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Version deleted successfully",
	})
}

func UpdateFileCurrentVersion(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	fileID, err := strconv.Atoi(vars["file_id"])
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	type RequestBody struct {
		VersionID int `json:"version_id"`
	}
	var reqBody RequestBody
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil || reqBody.VersionID == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// check the existence of the file and belonging to the user
	var ownerID int
	err = config.PostgresDB.QueryRow(`
		SELECT owner_id
		FROM Files
		WHERE file_id = $1
	`, fileID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if ownerID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// check file version
	var linkedFileID int
	err = config.PostgresDB.QueryRow(`
		SELECT file_id
		FROM FileVersions
		WHERE version_id = $1
	`, reqBody.VersionID).Scan(&linkedFileID)
	if err != nil {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}
	if linkedFileID != fileID {
		http.Error(w, "Version does not belong to the specified file", http.StatusBadRequest)
		return
	}

	// update current version
	_, err = config.PostgresDB.Exec(`
		UPDATE Files
		SET version_id = $1
		WHERE file_id = $2
	`, reqBody.VersionID, fileID)
	if err != nil {
		http.Error(w, "Failed to update current version", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Current version updated successfully",
	})
}
