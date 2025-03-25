package routes

import (
	"github.com/gorilla/mux"

	"backend/handlers"
	"backend/middleware"
)

func RegisterRoutes() *mux.Router {
	router := mux.NewRouter()

	// auth
	router.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	router.HandleFunc("/login", handlers.LoginUser).Methods("POST")

	// protect
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	// files
	protected.HandleFunc("/files/upload", handlers.UploadFile).Methods("POST")
	protected.HandleFunc("/files/{file_id}", handlers.DownloadFile).Methods("GET")
	protected.HandleFunc("/files/{file_id}", handlers.UpdateFile).Methods("PUT")
	protected.HandleFunc("/files/{file_id}", handlers.DeleteFile).Methods("DELETE")
	protected.HandleFunc("/filelist", handlers.GetUserFiles).Methods("GET")

	return router
}
