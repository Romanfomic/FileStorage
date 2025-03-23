package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"

	"backend/config"
	"backend/middleware"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(50 << 20)

	vars := mux.Vars(r)
	full_path, ok := vars["full_path"]
	if !ok || full_path == "" {
		full_path = "/"
	}

	fileType, ok := vars["type"]
	if !ok || fileType == "" {
		fileType = "file"
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

	bucket, err := gridfs.NewBucket(config.DB)
	if err != nil {
		http.Error(w, "GridFS initialization error", http.StatusInternalServerError)
		return
	}

	uploadStream, err := bucket.OpenUploadStream(
		header.Filename,
		options.GridFSUpload().SetMetadata(bson.M{
			"owner_id": userID,
		}),
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

	versionQuery := `
		INSERT INTO FileVersions (user_id, name, create_date, edit_date)
		VALUES ($1, $2, $3, $4)
		RETURNING version_id
	`
	versionName := "1.0"

	var versionID int
	err = config.PostgresDB.QueryRow(
		versionQuery,
		userID,      // user_id
		versionName, // name
		time.Now(),  // create_date
		time.Now(),  // edit_date
	).Scan(&versionID)

	if err != nil {
		http.Error(w, "Saving version error", http.StatusInternalServerError)
		return
	}

	query := `
		INSERT INTO Files (owner_id, group_id, role_id, access_id, version_id, create_date, edit_date, mongo_file_id, name, full_path, type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING file_id
	`

	mongoID := uploadStream.FileID.(primitive.ObjectID)
	mongoFileIDStr := mongoID.Hex()

	var fileID int
	err = config.PostgresDB.QueryRow(
		query,
		userID,          // owner_id
		sql.NullInt64{}, // group_id
		sql.NullInt64{}, // role_id
		sql.NullInt64{}, // access_id
		versionID,       // version_id
		time.Now(),      // create_date
		time.Now(),      // edit_date
		mongoFileIDStr,  // mongo_file_id as string
		header.Filename, // name
		full_path,       // full_path
		fileType,        // type
	).Scan(&fileID)

	if err != nil {
		http.Error(w, "Saving file metadata error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File %s uploaded", header.Filename)
}
