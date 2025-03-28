package handlers

import (
	"backend/config"
	"backend/middleware"
	"backend/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// required fields
	if user.Mail == "" || user.Login == "" || user.Password == "" || user.Name == "" || user.Surname == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	user.Type = "user"

	if err := user.HashPassword(); err != nil {
		http.Error(w, "Hash password error", http.StatusInternalServerError)
		return
	}

	query := `
		INSERT INTO Users (mail, login, password, name, surname, type, role_id, group_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING user_id
	`
	err := config.PostgresDB.QueryRow(query, user.Mail, user.Login, user.Password, user.Name, user.Surname, user.Type, user.RoleID, user.GroupID).Scan(&user.ID)
	if err != nil {
		http.Error(w, "Database insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user.ID)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var user models.User
	query := `
		SELECT user_id, mail, login, name, surname, type, role_id, group_id
		FROM Users
		WHERE user_id = $1
	`
	err := config.PostgresDB.QueryRow(query, userID).Scan(&user.ID, &user.Mail, &user.Login, &user.Name, &user.Surname, &user.Type, &user.RoleID, &user.GroupID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasPermission, err := middleware.CheckPermission(userID, "manage_users")
	if err != nil {
		http.Error(w, "Error checking permissions", http.StatusInternalServerError)
		return
	}
	if !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	groupIDParam := r.URL.Query().Get("group_id")
	var users []models.User
	var rows *sql.Rows

	if groupIDParam != "" {
		groupID, err := strconv.Atoi(groupIDParam)
		if err != nil {
			http.Error(w, "Invalid group_id", http.StatusBadRequest)
			return
		}

		rows, err = config.PostgresDB.Query(`
			SELECT user_id, login, mail, name, surname, type, role_id, group_id 
			FROM Users WHERE group_id = $1`, groupID)
	} else {
		rows, err = config.PostgresDB.Query(`
			SELECT user_id, login, mail, name, surname, type, role_id, group_id 
			FROM Users`)
	}

	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Login, &user.Mail, &user.Name, &user.Surname, &user.Type, &user.RoleID, &user.GroupID)
		if err != nil {
			http.Error(w, "Error scanning users", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	requesterID, _ := r.Context().Value(middleware.UserIDKey).(int)

	var updateData models.User
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var existingUser models.User
	err := config.PostgresDB.QueryRow("SELECT user_id, type FROM Users WHERE user_id = $1", requesterID).Scan(&existingUser.ID, &existingUser.Type)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	canManageUsers, err := middleware.CheckPermission(requesterID, "manage_users")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// required fields
	fields := []string{}
	values := []interface{}{}
	index := 1

	// user can change only fields: mail, login, password, name, surname
	if updateData.Mail != "" {
		fields = append(fields, fmt.Sprintf("mail = $%d", index))
		values = append(values, updateData.Mail)
		index++
	}
	if updateData.Login != "" {
		fields = append(fields, fmt.Sprintf("login = $%d", index))
		values = append(values, updateData.Login)
		index++
	}
	if updateData.Password != "" {
		fields = append(fields, fmt.Sprintf("password = $%d", index))
		values = append(values, updateData.Password)
		index++
	}
	if updateData.Name != "" {
		fields = append(fields, fmt.Sprintf("name = $%d", index))
		values = append(values, updateData.Name)
		index++
	}
	if updateData.Surname != "" {
		fields = append(fields, fmt.Sprintf("surname = $%d", index))
		values = append(values, updateData.Surname)
		index++
	}

	// admin and manage_users can change fields: role_id и group_id
	if updateData.RoleID != nil && canManageUsers {
		fields = append(fields, fmt.Sprintf("role_id = $%d", index))
		values = append(values, *updateData.RoleID)
		index++
	}
	if updateData.GroupID != nil && canManageUsers {
		fields = append(fields, fmt.Sprintf("group_id = $%d", index))
		values = append(values, *updateData.GroupID)
		index++
	}

	if len(fields) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("UPDATE Users SET %s WHERE user_id = $%d",
		strings.Join(fields, ", "), index)
	values = append(values, userID)

	_, err = config.PostgresDB.Exec(query, values...)
	if err != nil {
		http.Error(w, "Database update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	requesterID, _ := r.Context().Value(middleware.UserIDKey).(int)

	canManageUsers, err := middleware.CheckPermission(requesterID, "manage_users")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !canManageUsers {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = config.PostgresDB.Exec("DELETE FROM Users WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "Database delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
