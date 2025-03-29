package models

type FilePermissionUser struct {
	UserID   int `json:"user_id"`
	AccessID int `json:"access_id"`
}

type FilePermissionGroup struct {
	GroupID  int `json:"group_id"`
	AccessID int `json:"access_id"`
}
