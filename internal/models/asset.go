package models

import "gorm.io/gorm"

type Asset struct {
	gorm.Model
	Name      string `json:"name"`
	Type      string `json:"type"`
	Size      int64  `json:"size"`
	Path      string `json:"path"`
	Encrypted bool   `json:"encrypted"`
}
