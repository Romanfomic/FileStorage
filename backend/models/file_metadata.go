package models

type FileMetadata struct {
	FileID     int    `json:"file_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	FullPath   string `json:"full_path"`
	CreateDate string `json:"create_date"`
	EditDate   string `json:"edit_date"`
	VersionID  int    `json:"version_id"`
	OwnerID    *int   `json:"owner_id"`
}

type SharedFile struct {
	FileID     int    `json:"file_id"`
	Name       string `json:"name"`
	FullPath   string `json:"full_path"`
	OwnerID    int    `json:"owner_id"`
	GroupIDs   []int  `json:"group_ids,omitempty"`
	CreateDate string `json:"create_date"`
	EditDate   string `json:"edit_date"`
	VersionID  int    `json:"version_id"`
	AccessID   int    `json:"access_id"`
}

type FileVersion struct {
	VersionID  int    `json:"version_id"`
	Name       string `json:"name"`
	CreateDate string `json:"create_date"`
	EditDate   string `json:"edit_date"`
	IsCurrent  bool   `json:"is_current"`
}
