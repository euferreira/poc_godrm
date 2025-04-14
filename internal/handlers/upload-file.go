package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/models"
)

func UploadHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não foi enviado"})
		return
	}

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Não foi possível fechar o arquivo"})
			return
		}
	}(file)

	tempDir := "temp"
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
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

	dstPath := filepath.Join("temp", header.Filename)
	out, err := os.Create(dstPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Não foi possível criar o arquivo"})
		return
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Não foi possível fechar o arquivo"})
			return
		}
	}(out)

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar o arquivo"})
		return
	}

	asset := models.Asset{
		Name:      header.Filename,
		Type:      header.Header.Get("Content-Type"),
		Size:      header.Size,
		Path:      dstPath,
		Encrypted: false,
	}

	if err := database.DB.Create(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar metadados no banco"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Arquivo salvo com sucesso",
		"filename": header.Filename,
		"path":     dstPath,
		"size":     header.Size,
		"type":     header.Header.Get("Content-Type"),
	})
}
