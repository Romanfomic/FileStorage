package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"

	"backend/config"
)

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	// Получаем file_id из URL
	vars := mux.Vars(r)
	fileID, err := primitive.ObjectIDFromHex(vars["file_id"])
	if err != nil {
		http.Error(w, "Некорректный ID файла", http.StatusBadRequest)
		return
	}

	// Подключение к GridFS
	bucket, err := gridfs.NewBucket(config.DB)
	if err != nil {
		http.Error(w, "Ошибка доступа к GridFS", http.StatusInternalServerError)
		return
	}

	// Читаем файл из GridFS
	w.Header().Set("Content-Disposition", "attachment; filename="+vars["file_id"])
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = bucket.DownloadToStream(fileID, w) // Теперь мы явно игнорируем первый возвращаемый параметр
	if err != nil {
		http.Error(w, "Файл не найден", http.StatusNotFound)
	}
}
