package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/models"
	"projeto_drm/poc/internal/queue"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UploadHandler(c *gin.Context) {
	const maxUploadSize = 500 << 20 // 500 MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não foi enviado ou excede o limite de 500MB"})
		return
	}

	defer file.Close()

	allowedTypes(header.Header.Get("Content-Type"), c)

	if strings.Contains(header.Filename, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nome de arquivo inválido"})
		return
	}

	tempDir := "temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar diretório temporário"})
		return
	}

	var existingAsset models.Asset
	err = database.DB.Where("name = ?", header.Filename).First(&existingAsset).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// Trata erros inesperados do banco
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar existência do arquivo"})
		return
	}

	if err == nil {
		// Registro já existe
		c.JSON(http.StatusConflict, gin.H{"error": "Arquivo já existe"})
		return
	}

	safeName := filepath.Base(header.Filename)
	uniqueName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
	dstPath := filepath.Join(tempDir, uniqueName)

	if _, err := os.Stat(dstPath); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Arquivo já existe no sistema"})
		return
	}

	asset := models.Asset{
		Name:      header.Filename,
		Type:      header.Header.Get("Content-Type"),
		Size:      header.Size,
		Path:      dstPath,
		Status:    models.StatusPending,
		Encrypted: false,
	}

	if err := database.DB.Create(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar metadados no banco"})
		return
	}

	job := queue.AssetJob{
		ID:   asset.ID,
		Path: dstPath,
		Type: asset.Type,
	}

	if err := queue.EnqueueAssetJob(job); err != nil {
		database.DB.Model(&asset).Updates(models.Asset{Status: models.StatusFailed})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enfileirar processamento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Arquivo enviado com sucesso",
		"filename":     header.Filename,
		"path":         dstPath,
		"size":         header.Size,
		"type":         header.Header.Get("Content-Type"),
		"asset_id":     asset.ID,
		"download_url": fmt.Sprintf("/assets/%d/download", asset.ID),
		"status":       asset.Status,
	})
}

func allowedTypes(contentType string, c *gin.Context) {
	allowedTypes := map[string]bool{
		"application/pdf": true,
		"video/mp4":       true,
		"video/quicktime": true,
	}

	if !allowedTypes[contentType] {
		fmt.Println("Tipo de arquivo não suportado:", contentType)
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Tipo de arquivo não suportado"})
		c.Abort()
		return
	}
}
