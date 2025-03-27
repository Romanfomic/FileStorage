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
	protected.HandleFunc("/files", handlers.GetUserFiles).Methods("GET")

	// roles
	protected.HandleFunc("/roles", middleware.RequirePermission("manage_roles", handlers.CreateRole)).Methods("POST")
	protected.HandleFunc("/roles/{id}", middleware.RequirePermission("manage_roles", handlers.GetRole)).Methods("GET")
	protected.HandleFunc("/roles", middleware.RequirePermission("manage_roles", handlers.GetRoles)).Methods("GET")
	protected.HandleFunc("/roles/{id}", middleware.RequirePermission("manage_roles", handlers.UpdateRole)).Methods("PUT")
	protected.HandleFunc("/roles", middleware.RequirePermission("manage_roles", handlers.DeleteRole)).Methods("DELETE")

	// groups
	protected.HandleFunc("/groups", middleware.RequirePermission("manage_groups", handlers.CreateGroup)).Methods("POST")
	protected.HandleFunc("/groups", middleware.RequirePermission("manage_groups", handlers.GetGroups)).Methods("GET")
	protected.HandleFunc("/groups/{id}", middleware.RequirePermission("manage_groups", handlers.UpdateGroup)).Methods("PUT")
	protected.HandleFunc("/groups/{id}", middleware.RequirePermission("manage_groups", handlers.DeleteGroup)).Methods("DELETE")

	return router
}
