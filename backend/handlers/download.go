package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"

	"backend/config"
	"backend/middleware"
)

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Не удалось получить ID пользователя", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	fileID, err := primitive.ObjectIDFromHex(vars["file_id"])
	if err != nil {
		http.Error(w, "Некорректный ID файла", http.StatusBadRequest)
		return
	}

	bucket, err := gridfs.NewBucket(config.DB)
	if err != nil {
		http.Error(w, "Ошибка доступа к GridFS", http.StatusInternalServerError)
		return
	}

	var fileMetadata struct {
		Filename string `bson:"filename"`
		Metadata struct {
			OwnerID string `bson:"owner_id"`
		} `bson:"metadata"`
	}

	err = config.DB.Collection("fs.files").FindOne(context.TODO(), bson.M{"_id": fileID}).Decode(&fileMetadata)
	if err != nil {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}

	if fileMetadata.Metadata.OwnerID != userID {
		http.Error(w, "У вас нет доступа к этому файлу", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileMetadata.Filename+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = bucket.DownloadToStream(fileID, w)
	if err != nil {
		http.Error(w, "Ошибка при загрузке файла", http.StatusInternalServerError)
	}
}
