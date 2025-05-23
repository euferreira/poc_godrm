package models

import (
	"gorm.io/gorm"
	"time"
)

type Asset struct {
	gorm.Model
	Name      string `json:"name"`
	Path      string `json:"path"`
	Type      string `json:"type"`
	Size      int64  `json:"size"`
	Status    string `json:"status" gorm:"default:pending"`
	Encrypted bool   `json:"encrypted"`
}

type ProcessedAsset struct {
	gorm.Model
	AssetID     uint       `json:"asset_id" gorm:"index"`
	UserID      uint       `json:"user_id" gorm:"index"`
	Status      string     `json:"status" gorm:"default:queued"` // "queued", "processing", "completed", "failed"
	CachePath   string     `json:"cache_path"`
	ProcessedAt *time.Time `json:"processed_at"`
	ErrorMsg    string     `json:"error_msg,omitempty"`
}

func (ProcessedAsset) TableName() string {
	return "processed_assets"
}
