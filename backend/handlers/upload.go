package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/gridfs"

	"backend/config"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	// Ограничение размера файла (50MB)
	r.ParseMultipartForm(50 << 20)

	// Получение файла из запроса
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Print(err)
		http.Error(w, "Ошибка загрузки файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Подключение к GridFS
	bucket, err := gridfs.NewBucket(config.DB)
	if err != nil {
		http.Error(w, "Ошибка инициализации GridFS", http.StatusInternalServerError)
		return
	}

	// Запись файла в GridFS
	uploadStream, err := bucket.OpenUploadStream(header.Filename)
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

	// Сохранение метаданных в MongoDB
	fileID := uploadStream.FileID
	metadata := bson.M{
		"filename": header.Filename,
		"size":     header.Size,
		"uploaded": time.Now(),
	}

	_, err = config.DB.Collection("fs.files").UpdateOne(context.TODO(),
		bson.M{"_id": fileID}, bson.M{"$set": metadata})
	if err != nil {
		log.Println("Ошибка сохранения метаданных:", err)
	}

	fmt.Fprintf(w, "✅ Файл %s загружен! ID: %v", header.Filename, fileID)
}
