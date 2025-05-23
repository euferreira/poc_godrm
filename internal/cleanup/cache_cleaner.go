package cleanup

import (
	"log"
	"os"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/models"
	"time"
)

func StartCacheCleanup(interval time.Duration, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				cleanupOldCache(maxAge)
			}
		}
	}()
}

func cleanupOldCache(maxAge time.Duration) {
	log.Println("Starting cache cleanup...")

	cutoff := time.Now().Add(-maxAge)

	var processedAssets []models.ProcessedAsset
	err := database.DB.Where("status = ? AND processed_at < ?", "completed", cutoff).Find(&processedAssets).Error
	if err != nil {
		log.Printf("Error finding old processed assets: %v", err)
		return
	}

	cleaned := 0
	for _, pa := range processedAssets {
		if pa.CachePath != "" {
			if err := os.Remove(pa.CachePath); err != nil && !os.IsNotExist(err) {
				log.Printf("Error removing cache file %s: %v", pa.CachePath, err)
				continue
			}
		}

		// Remover registro do banco
		if err := database.DB.Delete(&pa).Error; err != nil {
			log.Printf("Error deleting processed asset record: %v", err)
			continue
		}

		cleaned++
	}

	log.Printf("Cache cleanup completed. Removed %d files", cleaned)
}
