package handlers

import (
	"github.com/gin-gonic/gin"
	"os"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/assets", ListAssets)
	r.GET("/assets/all", GetAllProcessStatus)
	r.GET("/assets/temp", func(c *gin.Context) {
		//verifica se existe o diretorio temp/
		if _, err := os.Stat("temp/"); os.IsNotExist(err) {
			//se não existe, retorna erro
			c.JSON(500, gin.H{"error": "Diretório temp não encontrado"})
		}

		//lê os arquivos dentro da pasta temp/
		files, err := os.ReadDir("temp/")
		if err != nil {
			c.JSON(500, gin.H{"error": "Erro ao ler diretório temp"})
			return
		}
		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, file.Name())
		}

		c.JSON(200, gin.H{"files": fileNames})
	})

	r.GET("/assets/:id", GetAsset)
	r.POST("/assets/:id/download", DownloadHandlerV2)
	r.GET("/assets/:id/status", CheckProcessingStatus)

	r.POST("/upload", UploadHandler)
	r.GET("/im-alive", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "I'm alive"})
	})
}
