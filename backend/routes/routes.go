package routes

import (
	"github.com/gorilla/mux"

	"backend/handlers"
)

func RegisterRoutes() *mux.Router {
	r := mux.NewRouter()

	// Эндпоинты
	r.HandleFunc("/files/upload", handlers.UploadFile).Methods("POST")
	r.HandleFunc("/files/{file_id}", handlers.DownloadFile).Methods("GET")

	return r
}
