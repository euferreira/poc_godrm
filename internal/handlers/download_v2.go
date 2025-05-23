package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"projeto_drm/poc/internal/auth"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/models"
	"projeto_drm/poc/internal/queue"
	"strconv"
	"time"
)

var redisQueue *queue.RedisQueue

func InitializeQueue(redisURL string) {
	redisQueue = queue.NewRedisQueue(redisURL)
}

func DownloadHandlerV2(c *gin.Context) {
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

	// Validar extensão
	ext := filepath.Ext(asset.Path)
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

	assetIDUint, err := strconv.ParseUint(assetID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do asset inválido"})
		return
	}
	userIDUint, err := strconv.ParseUint(user.ID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do usuário inválido"})
		return
	}

	var processedAsset models.ProcessedAsset
	erro := database.DB.Where("asset_id = ? AND user_id = ?", assetIDUint, userIDUint).First(&processedAsset).Error

	if erro == nil {
		// Já existe registro
		switch processedAsset.Status {
		case "completed":
			// Verificar se arquivo ainda existe no cache
			if _, err := os.Stat(processedAsset.CachePath); err == nil {
				filename := filepath.Base(asset.Path)
				c.FileAttachment(processedAsset.CachePath, filename)
				return
			} else {
				// Cache foi removido, reprocessar
				processedAsset.Status = "queued"
				processedAsset.CachePath = ""
				processedAsset.ProcessedAt = nil
				database.DB.Save(&processedAsset)
			}
		case "processing":
			c.JSON(http.StatusAccepted, gin.H{
				"status":  "processing",
				"message": "Arquivo sendo processado. Tente novamente em alguns instantes.",
				"job_id":  fmt.Sprintf("%s_%s", assetID, user.ID),
			})
			return
		case "queued":
			c.JSON(http.StatusAccepted, gin.H{
				"status":  "queued",
				"message": "Arquivo na fila de processamento. Tente novamente em alguns instantes.",
				"job_id":  fmt.Sprintf("%s_%s", assetID, user.ID),
			})
			return
		case "failed":
			// Tentar reprocessar
			processedAsset.Status = "queued"
			processedAsset.ErrorMsg = ""
			database.DB.Save(&processedAsset)
		}
	} else {
		// Criar novo registro
		processedAsset = models.ProcessedAsset{
			AssetID: uint(assetIDUint),
			UserID:  uint(userIDUint),
			Status:  "queued",
		}
		if err := database.DB.Create(&processedAsset).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar registro de processamento"})
			return
		}
	}

	// Enfileirar job
	jobID := uuid.New().String()
	job := queue.ProcessingJob{
		ID:        jobID,
		AssetID:   assetID,
		UserID:    user.ID,
		AssetPath: asset.Path,
		AssetType: ext,
		UserEmail: user.Email,
		CreatedAt: time.Now(),
	}

	if err := redisQueue.EnqueueJob(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enfileirar processamento"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "queued",
		"message": "Arquivo adicionado à fila de processamento. Tente novamente em alguns instantes.",
		"job_id":  fmt.Sprintf("%s_%s", assetID, user.ID),
	})
}

func CheckProcessingStatus(c *gin.Context) {
	log.Println("CheckProcessingStatus called")
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}
	user := userRaw.(auth.UserInfo)

	assetID := c.Param("id")

	log.Println("Buscando status do processamento para assetID:", assetID, "e userID:", user.ID)
	var processedAsset models.ProcessedAsset
	err := database.DB.Where("asset_id = ? AND user_id = ?", assetID, user.ID).First(&processedAsset).Error

	if err != nil {
		log.Println("Erro ao buscar status do processamento:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Processamento não encontrado"})
		return
	}

	response := gin.H{
		"asset_id":     assetID,
		"status":       processedAsset.Status,
		"processed_at": processedAsset.ProcessedAt,
	}

	if processedAsset.ErrorMsg != "" {
		response["error"] = processedAsset.ErrorMsg
	}

	c.JSON(http.StatusOK, response)
}

func GetAllProcessStatus(c *gin.Context) {
	userRaw, exists := c.Get("user")
	log.Println("GetAllProcessStatus called", userRaw)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	//recupera os status da tabela processed_assets, todos eles, faz um select * from
	var processedAssets []models.ProcessedAsset
	err := database.DB.Find(&processedAssets).Error

	if err != nil {
		log.Println("Erro ao buscar status do processamento:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar status do processamento"})
		return
	}

	c.JSON(http.StatusOK, processedAssets)
}
