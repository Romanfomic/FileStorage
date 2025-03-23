package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"

	"backend/config"
)

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["file_id"]

	// Получаем mongo_file_id из PostgreSQL
	var mongoFileIDStr string

	err := config.PostgresDB.QueryRow(`
		SELECT mongo_file_id
		FROM Files
		WHERE file_id = $1
	`, fileID).Scan(&mongoFileIDStr)

	if err == sql.ErrNoRows {
		http.Error(w, "Файл не найден в PostgreSQL", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Ошибка чтения из PostgreSQL", http.StatusInternalServerError)
		return
	}

	// Преобразуем строку mongoFileID в ObjectID
	mongoFileID, err := primitive.ObjectIDFromHex(mongoFileIDStr)
	if err != nil {
		http.Error(w, "Некорректный mongo_file_id", http.StatusInternalServerError)
		return
	}

	// Загружаем файл из GridFS
	bucket, err := gridfs.NewBucket(config.DB)
	if err != nil {
		http.Error(w, "Ошибка инициализации GridFS", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = bucket.DownloadToStream(mongoFileID, w)
	if err != nil {
		http.Error(w, "Ошибка загрузки файла из GridFS", http.StatusInternalServerError)
		return
	}
}
