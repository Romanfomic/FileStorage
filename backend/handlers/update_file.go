package handlers

import (
	"backend/config"
	"backend/middleware"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

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
