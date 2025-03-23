package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"backend/config"
	"backend/models"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secret_key")

type CustomClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "JSON decode error", http.StatusBadRequest)
		return
	}

	var existingID int
	err := config.PostgresDB.QueryRow("SELECT user_id FROM Users WHERE login = $1", user.Login).Scan(&existingID)
	if err != sql.ErrNoRows {
		http.Error(w, "Login is already exist", http.StatusConflict)
		return
	}

	err = config.PostgresDB.QueryRow("SELECT user_id FROM Users WHERE mail = $1", user.Mail).Scan(&existingID)
	if err != sql.ErrNoRows {
		http.Error(w, "Mail is already exist", http.StatusConflict)
		return
	}

	if err := user.HashPassword(); err != nil {
		http.Error(w, "Hash password error", http.StatusInternalServerError)
		return
	}

	var userID int
	err = config.PostgresDB.QueryRow(`
		INSERT INTO Users (login, password, mail, name, surname)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING user_id
	`, user.Login, user.Password, user.Mail, user.Name, user.Surname).Scan(&userID)

	if err != nil {
		http.Error(w, "Create new user error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User created successfully",
		"user_id": userID,
	})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var creds models.User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "JSON decode error", http.StatusBadRequest)
		return
	}

	var userID int
	var hashedPassword string
	err := config.PostgresDB.QueryRow(`
		SELECT user_id, password FROM Users WHERE login = $1
	`, creds.Login).Scan(&userID, &hashedPassword)

	if err == sql.ErrNoRows || !models.CheckPasswordHash(creds.Password, hashedPassword) {
		http.Error(w, "Incorrect login or password", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Login error", http.StatusInternalServerError)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Token generation error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}
