package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/models"
)

func processFile(asset models.Asset, file multipart.File, dstPath string) {
	// Atualizar status para "processing"
	asset.Status = models.StatusProcessing
	if err := database.DB.Save(&asset).Error; err != nil {
		fmt.Printf("Erro ao atualizar status para 'processing': %v\n", err)
		return
	}

	// Fechar o arquivo após o uso
	defer func() {
		if err := file.Close(); err != nil {
			updateAssetStatus(&asset, models.StatusFailed, fmt.Sprintf("Erro ao fechar o arquivo: %v", err))
		}
	}()

	// Criar o arquivo de destino
	out, err := os.Create(dstPath)
	if err != nil {
		updateAssetStatus(&asset, models.StatusFailed, fmt.Sprintf("Erro ao criar arquivo de destino: %v", err))
		return
	}
	defer out.Close()

	// Copiar o conteúdo do arquivo
	if _, err = io.Copy(out, file); err != nil {
		updateAssetStatus(&asset, models.StatusFailed, fmt.Sprintf("Erro ao copiar arquivo: %v", err))
		return
	}

	// Atualizar status para "completed"
	if err := updateAssetStatus(&asset, models.StatusCompleted, ""); err != nil {
		fmt.Printf("Erro ao atualizar status para 'completed': %v\n", err)
	}
}

// Função auxiliar para atualizar o status do asset
func updateAssetStatus(asset *models.Asset, status string, errorMessage string) error {
	asset.Status = status
	if err := database.DB.Save(asset).Error; err != nil {
		fmt.Printf("Erro ao atualizar status para '%s': %v\n", status, err)
		return err
	}
	if errorMessage != "" {
		fmt.Println(errorMessage)
	}
	return nil
}
