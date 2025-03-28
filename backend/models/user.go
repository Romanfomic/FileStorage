package models

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"user_id"`
	Mail     string `json:"mail"`
	Login    string `json:"login"`
	Password string `json:"password,omitempty"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Type     string `json:"type"`
	RoleID   *int   `json:"role_id"`
	GroupID  *int   `json:"group_id"`
}

func (u *User) HashPassword() error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
