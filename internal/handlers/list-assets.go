package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/models"
)

type ListAssetsResponse struct {
	Assets []models.Asset `json:"assets"`
}

func ListAssets(c *gin.Context) {
	var assets []models.Asset

	if err := database.DB.Find(&assets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve assets",
		})
	}

	c.JSON(http.StatusOK, ListAssetsResponse{
		Assets: assets,
	})
}

func GetAsset(c *gin.Context) {
	assetID := c.Param("id")
	var asset models.Asset

	if err := database.DB.First(&asset, assetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Asset not found",
		})
		return
	}

	c.JSON(http.StatusOK, asset)
}
