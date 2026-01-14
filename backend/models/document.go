package models

import "time"

type Document struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Content   string    `gorm:"type:longtext" json:"content"`
	OwnerID   uint      `json:"owner_id"`
	IsStarred bool      `json:"is_starred"`
	FolderID  int       `json:"folder_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DocVersion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DocID       uint      `json:"doc_id"`
	Content     string    `gorm:"type:longtext" json:"content"`
	VersionName string    `json:"version_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type Folder struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Name   string `json:"name"`
	UserID uint   `json:"user_id"`
}

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DocID     uint      `json:"doc_id"`
	UserID    uint      `json:"user_id"`
	Content   string    `json:"content"`
	LineNum   int       `json:"line_num"`
	CreatedAt time.Time `json:"created_at"`
}
