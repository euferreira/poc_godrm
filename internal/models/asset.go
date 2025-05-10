package models

import "gorm.io/gorm"

type Asset struct {
	gorm.Model
	Name      string `json:"name"`
	Path      string `json:"path"`
	Type      string `json:"type"`
	Size      int64  `json:"size"`
	Status    string `json:"status" gorm:"default:pending"`
	Encrypted bool   `json:"encrypted"`
}
