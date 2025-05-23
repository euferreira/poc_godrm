package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
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
	const maxUploadSize = 1 << 30 // 1GB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não foi enviado ou excede o limite de 1GB"})
		return
	}
	defer file.Close()

	allowedTypes(header.Header.Get("Content-Type"), c)
	if strings.Contains(header.Filename, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nome de arquivo inválido"})
		return
	}

	var existingAsset models.Asset
	err = database.DB.Where("name = ?", header.Filename).First(&existingAsset).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar existência do arquivo"})
		return
	}
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Arquivo já existe"})
		return
	}

	storageType := os.Getenv("STORAGE_TYPE")
	log.Println("Tipo de armazenamento:", storageType)
	if storageType == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo de armazenamento não definido"})
		return
	}

	switch storageType {
	case "local":
		uploadLocalFile(c, file, header)
		break
	case "s3":
		uploadS3File(c, file, header)
		break

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo de armazenamento não suportado"})
		return
	}
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

func uploadLocalFile(c *gin.Context, file io.Reader, header *multipart.FileHeader) {
	tempDir := "temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar diretório temporário"})
		return
	}

	safeName := filepath.Base(header.Filename)
	uniqueName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
	dstPath := filepath.Join(tempDir, uniqueName)

	if _, err := os.Stat(dstPath); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Arquivo já existe no sistema"})
		return
	}

	// Create a temporary file to store the uploaded content
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar arquivo temporário"})
		return
	}
	tempFilePath := tempFile.Name()
	defer tempFile.Close()

	// Copy the uploaded file to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		os.Remove(tempFilePath) // Clean up the temporary file
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo temporário"})
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
		os.Remove(tempFilePath) // Clean up the temporary file
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar metadados no banco"})
		return
	}

	job := queue.AssetJob{
		ID:           asset.ID,
		Path:         dstPath,
		Type:         asset.Type,
		TempFilePath: tempFilePath,
	}

	if err := queue.EnqueueAssetJob(job); err != nil {
		os.Remove(tempFilePath) // Clean up the temporary file
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

func uploadS3File(c *gin.Context, file io.Reader, header *multipart.FileHeader) {
	// Implementar upload para S3
	c.JSON(http.StatusOK, gin.H{
		"message": "Upload para S3 não implementado",
	})
}
