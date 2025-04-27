package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"backend/handlers"
	"backend/middleware"
)

func RegisterRoutes() http.Handler {
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

	// permissions
	protected.HandleFunc("/permissions", middleware.RequirePermission("manage_roles", handlers.GetPermissions)).Methods("GET")
	protected.HandleFunc("/permissions/{id}", middleware.RequirePermission("manage_roles", handlers.GetPermission)).Methods("GET")

	// groups
	protected.HandleFunc("/groups", middleware.RequirePermission("manage_groups", handlers.CreateGroup)).Methods("POST")
	protected.HandleFunc("/groups", middleware.RequirePermission("manage_groups", handlers.GetGroups)).Methods("GET")
	protected.HandleFunc("/groups/{id}", middleware.RequirePermission("manage_groups", handlers.UpdateGroup)).Methods("PUT")
	protected.HandleFunc("/groups/{id}", middleware.RequirePermission("manage_groups", handlers.DeleteGroup)).Methods("DELETE")

	// users
	protected.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	protected.HandleFunc("/users/{id}", handlers.GetUser).Methods("GET")
	protected.HandleFunc("/users", handlers.GetUsers).Methods("GET")
	protected.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT")
	protected.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE")

	// sharing
	protected.HandleFunc("/files/{file_id}/share/user", handlers.ShareFileWithUser).Methods("POST")
	protected.HandleFunc("/files/{file_id}/share/group", handlers.ShareFileWithGroup).Methods("POST")
	protected.HandleFunc("/files/{file_id}/permissions", handlers.GetFilePermissions).Methods("GET")
	protected.HandleFunc("/shared-files", handlers.GetSharedFiles).Methods("GET")
	protected.HandleFunc("/files/{file_id}/share/user/{user_id}", handlers.RevokeUserAccess).Methods("DELETE")
	protected.HandleFunc("/files/{file_id}/share/group/{group_id}", handlers.RevokeGroupAccess).Methods("DELETE")

	// versions
	protected.HandleFunc("/files/{file_id}/version", handlers.CreateFileVersion).Methods("POST")
	protected.HandleFunc("/files/{file_id}/versions", handlers.GetFileVersions).Methods("GET")
	protected.HandleFunc("/files/{file_id}/version", handlers.UpdateFileCurrentVersion).Methods("PUT")
	protected.HandleFunc("/versions/{version_id}", handlers.UpdateFileVersionName).Methods("PUT")
	protected.HandleFunc("/versions/{version_id}", handlers.DeleteFileVersion).Methods("DELETE")

	return middleware.CORSMiddleware(router)
}
