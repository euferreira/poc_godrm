package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"projeto_drm/poc/internal/auth"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/models"
	"projeto_drm/poc/internal/watermarker"
)

func DownloadHandler(c *gin.Context) {
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}
	user := userRaw.(auth.UserInfo)

	assetID := c.Param("id")
	var asset models.Asset
	if err := database.DB.First(&asset, assetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo não encontrado"})
		return
	}

	ext := filepath.Ext(asset.Path)
	filename := filepath.Base(asset.Path)
	outputPath := filepath.Join("temp", fmt.Sprintf("%s_%s", user.ID, filename))

	var err error
	switch ext {
	case ".pdf":
		err = watermarker.AddPDFWatermark(asset.Path, outputPath, fmt.Sprintf("%s (%s)", user.ID, user.Email))
	case ".mp4", ".mov":
		err = watermarker.AddVideoWatermark(asset.Path, outputPath, fmt.Sprintf("%s (%s)", user.ID, user.Email))
	default:
		// Arquivo sem watermarking
		outputPath = asset.Path
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao aplicar marca: %v", err)})
		return
	}

	c.FileAttachment(outputPath, filename)
}
