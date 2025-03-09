package handlers

import (
	"fmt"
	"io"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"

	"backend/config"
	"backend/middleware"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	// Ограничение размера файла (50MB)
	r.ParseMultipartForm(50 << 20)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка загрузки файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Не удалось получить ID пользователя", http.StatusUnauthorized)
		return
	}

	bucket, err := gridfs.NewBucket(config.DB)
	if err != nil {
		http.Error(w, "Ошибка инициализации GridFS", http.StatusInternalServerError)
		return
	}

	uploadStream, err := bucket.OpenUploadStream(
		header.Filename,
		options.GridFSUpload().SetMetadata(bson.M{
			"owner_id": userID,
		}),
	)
	if err != nil {
		http.Error(w, "Ошибка записи файла", http.StatusInternalServerError)
		return
	}
	defer uploadStream.Close()

	_, err = io.Copy(uploadStream, file)
	if err != nil {
		http.Error(w, "Ошибка сохранения файла", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "✅ Файл %s загружен! ID: %v", header.Filename, uploadStream.FileID)
}
