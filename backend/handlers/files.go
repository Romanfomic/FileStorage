package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"

	"backend/config"
	"backend/middleware"
	"backend/models"
)

/*
form-data file: file
*/
func UploadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File upload error", http.StatusBadRequest)
		return
	}
	defer file.Close()

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusUnauthorized)
		return
	}

	fullPath := r.FormValue("full_path")
	if fullPath == "" {
		fullPath = "/"
	}

	fileType := r.FormValue("type")
	if fileType == "" {
		fileType = "file"
	}

	bucket, err := gridfs.NewBucket(config.DB)
	if err != nil {
		http.Error(w, "GridFS initialization error", http.StatusInternalServerError)
		return
	}

	uploadStream, err := bucket.OpenUploadStream(
		header.Filename,
		options.GridFSUpload().SetMetadata(bson.M{"owner_id": userID}),
	)
	if err != nil {
		http.Error(w, "Saving file error", http.StatusInternalServerError)
		return
	}
	defer uploadStream.Close()

	_, err = io.Copy(uploadStream, file)
	if err != nil {
		http.Error(w, "File write error", http.StatusInternalServerError)
		return
	}

	mongoID := uploadStream.FileID.(primitive.ObjectID)
	mongoFileIDStr := mongoID.Hex()

	var versionID int
	err = config.PostgresDB.QueryRow(`
		INSERT INTO FileVersions (user_id, name)
		VALUES ($1, $2)
		RETURNING version_id
	`, userID, "1.0").Scan(&versionID)

	if err != nil {
		http.Error(w, "Saving version error", http.StatusInternalServerError)
		return
	}

	var fileID int
	err = config.PostgresDB.QueryRow(`
		INSERT INTO Files (owner_id, version_id, mongo_file_id, name, full_path, type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING file_id
	`, userID, versionID, mongoFileIDStr, header.Filename, fullPath, fileType).Scan(&fileID)

	if err != nil {
		http.Error(w, "Saving file metadata error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "File uploaded successfully",
		"file_id": fileID,
	})
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["file_id"]

	var mongoFileIDStr, fileName, fileType string

	err := config.PostgresDB.QueryRow(`
		SELECT mongo_file_id, name, type
		FROM Files
		WHERE file_id = $1
	`, fileID).Scan(&mongoFileIDStr, &fileName, &fileType)

	if err == sql.ErrNoRows {
		http.Error(w, "The file was not found in PostgreSQL", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error reading from PostgreSQL", http.StatusInternalServerError)
		return
	}

	mongoFileID, err := primitive.ObjectIDFromHex(mongoFileIDStr)
	if err != nil {
		http.Error(w, "Invalid mongo_file_id", http.StatusInternalServerError)
		return
	}

	bucket, err := gridfs.NewBucket(config.DB)
	if err != nil {
		http.Error(w, "GridFS initialization error", http.StatusInternalServerError)
		return
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	case ".mp4":
		w.Header().Set("Content-Type", "video/mp4")
	case ".mp3":
		w.Header().Set("Content-Type", "audio/mpeg")
	case ".wav":
		w.Header().Set("Content-Type", "audio/wav")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileName))

	_, err = bucket.DownloadToStream(mongoFileID, w)
	if err != nil {
		http.Error(w, "Error downloading a file from GridFS", http.StatusInternalServerError)
		return
	}
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
			version_id, owner_id
		FROM Files
		WHERE owner_id = $1
	`

	rows, err := config.PostgresDB.Query(query, userID)
	if err != nil {
		http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []models.FileMetadata
	for rows.Next() {
		var file models.FileMetadata
		err := rows.Scan(
			&file.FileID,
			&file.Name,
			&file.Type,
			&file.FullPath,
			&file.CreateDate,
			&file.EditDate,
			&file.VersionID,
			&file.OwnerID,
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

/*
form-data file: file | name: string
*/
func UpdateFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fileID := params["file_id"]
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusUnauthorized)
		return
	}

	// Check owner
	var ownerID int
	var mongoFileID, fileName string
	err := config.PostgresDB.QueryRow("SELECT owner_id, mongo_file_id, name FROM Files WHERE file_id = $1", fileID).Scan(&ownerID, &mongoFileID, &fileName)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if ownerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var requestData struct {
		Name string `json:"name"`
	}
	if r.Header.Get("Content-Type") == "application/json" {
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	}
	newFileName := requestData.Name
	if newFileName == "" {
		newFileName = fileName
	}

	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()

		// Change filename if exist
		if header.Filename != "" {
			newFileName = header.Filename
		}

		// Delete old file in mongo
		collection := config.DB.Collection("fs.files")
		_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": mongoFileID})
		if err != nil {
			http.Error(w, "Failed to delete old file from storage", http.StatusInternalServerError)
			return
		}

		// Load new file in mongo
		bucket, err := gridfs.NewBucket(config.DB)
		if err != nil {
			http.Error(w, "Failed to get GridFS bucket", http.StatusInternalServerError)
			return
		}
		uploadStream, err := bucket.OpenUploadStream(fmt.Sprintf("file_%s", fileID))
		if err != nil {
			http.Error(w, "Failed to upload file", http.StatusInternalServerError)
			return
		}
		defer uploadStream.Close()
		_, err = io.Copy(uploadStream, file)
		if err != nil {
			http.Error(w, "Failed to write file to storage", http.StatusInternalServerError)
			return
		}
		newMongoFileID := uploadStream.FileID.(primitive.ObjectID).Hex()

		_, err = config.PostgresDB.Exec("UPDATE Files SET mongo_file_id = $1, name = $2, edit_date = NOW() WHERE file_id = $3", newMongoFileID, newFileName, fileID)
		if err != nil {
			http.Error(w, "Failed to update file metadata", http.StatusInternalServerError)
			return
		}
	} else if newFileName != fileName {
		_, err = config.PostgresDB.Exec("UPDATE Files SET name = $1, edit_date = NOW() WHERE file_id = $2", newFileName, fileID)
		if err != nil {
			http.Error(w, "Failed to update file name", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "File updated"})
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fileID := params["file_id"]
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusUnauthorized)
		return
	}

	// Check owner
	var ownerID int
	var mongoFileID string
	err := config.PostgresDB.QueryRow("SELECT owner_id, mongo_file_id FROM Files WHERE file_id = $1", fileID).Scan(&ownerID, &mongoFileID)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if ownerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	collection := config.DB.Collection("fs.files")
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": mongoFileID})
	if err != nil {
		http.Error(w, "Failed to delete file from storage", http.StatusInternalServerError)
		return
	}

	_, err = config.PostgresDB.Exec("DELETE FROM Files WHERE file_id = $1", fileID)
	if err != nil {
		http.Error(w, "Failed to delete file from DB", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "File deleted"})
}
