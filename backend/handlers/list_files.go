package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"

	"backend/config"
	"backend/middleware"
)

func GetUserFiles(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Не удалось получить ID пользователя", http.StatusUnauthorized)
		return
	}

	cursor, err := config.DB.Collection("fs.files").Find(
		context.TODO(),
		bson.M{"metadata.owner_id": userID},
	)
	if err != nil {
		http.Error(w, "Ошибка поиска файлов", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var files []struct {
		FileID   interface{} `bson:"_id"`
		Filename string      `bson:"filename"`
	}
	for cursor.Next(context.TODO()) {
		var file struct {
			FileID   interface{} `bson:"_id"`
			Filename string      `bson:"filename"`
		}
		if err := cursor.Decode(&file); err != nil {
			http.Error(w, "Ошибка декодирования данных файла", http.StatusInternalServerError)
			return
		}
		files = append(files, file)
	}

	if len(files) == 0 {
		http.Error(w, "Нет файлов для данного пользователя", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(files)
	if err != nil {
		http.Error(w, "Ошибка при отправке данных", http.StatusInternalServerError)
		return
	}
}
