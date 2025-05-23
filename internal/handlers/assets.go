package handlers

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/assets", ListAssets)
	r.GET("/assets/process-status", GetAllProcessStatus)

	r.GET("/assets/:id", GetAsset)
	r.GET("/assets/:id/status", CheckProcessingStatus)
	r.POST("/assets/:id/download", DownloadHandlerV2)

	r.POST("/upload", UploadHandler)
	r.GET("/im-alive", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "I'm alive"})
	})
}
