package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"projeto_drm/poc/internal/auth"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/models"
	"projeto_drm/poc/internal/watermarker"
	"time"

	"github.com/gin-gonic/gin"
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
	allowedExtensions := []string{".pdf", ".mp4", ".mov"}
	isValidExtension := false
	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			isValidExtension = true
			break
		}
	}
	if !isValidExtension {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de arquivo não suportado"})
		return
	}

	cacheDir := "cache"
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar diretório de cache"})
		return
	}

	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%s_%s", user.ID, filename))
	if _, err := os.Stat(cachePath); err == nil {
		// Arquivo já processado, usar o cache
		c.FileAttachment(cachePath, filename)
		return
	}

	// Garantir que o diretório temporário exista
	tempDir := "temp"
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar diretório temporário"})
		return
	}

	// Gerar caminho único para o arquivo processado
	outputPath := filepath.Join("temp", fmt.Sprintf("%s_%d_%s", user.ID, time.Now().UnixNano(), filename))

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

	// Remover arquivo temporário após o envio
	defer os.Remove(outputPath)

	c.FileAttachment(outputPath, filename)
}
