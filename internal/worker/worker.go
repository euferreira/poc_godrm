package worker

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/models"
	"projeto_drm/poc/internal/queue"
	"projeto_drm/poc/internal/watermarker"
	"sync"
	"time"
)

type Worker struct {
	ID    int
	queue *queue.RedisQueue
	quit  chan bool
}

type WorkerPool struct {
	workers []*Worker
	queue   *queue.RedisQueue
	wg      sync.WaitGroup
}

func NewWorkerPool(size int, redisQueue *queue.RedisQueue) *WorkerPool {
	workers := make([]*Worker, size)
	for i := 0; i < size; i++ {
		workers[i] = &Worker{
			ID:    i + 1,
			queue: redisQueue,
			quit:  make(chan bool),
		}
	}

	return &WorkerPool{
		workers: workers,
		queue:   redisQueue,
	}
}

func (wp *WorkerPool) Start() {
	log.Printf("Starting worker pool with %d workers", len(wp.workers))

	for _, worker := range wp.workers {
		wp.wg.Add(1)
		go worker.Start(&wp.wg)
	}
}

func (wp *WorkerPool) Stop() {
	log.Println("Stopping worker pool...")

	for _, worker := range wp.workers {
		worker.quit <- true
	}

	wp.wg.Wait()
	log.Println("Worker pool stopped")
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("Worker %d started", w.ID)

	for {
		select {
		case <-w.quit:
			log.Printf("Worker %d stopping", w.ID)
			return
		default:
			job, err := w.queue.DequeueJob()
			if err != nil {
				log.Printf("Worker %d: Error dequeueing job: %v", w.ID, err)
				time.Sleep(time.Second)
				continue
			}

			w.processJob(job)
		}
	}
}

func (w *Worker) processJob(job *queue.ProcessingJob) {
	log.Printf("Worker %d: Processing job %s for user %s", w.ID, job.ID, job.UserID)

	// Atualizar status para "processing"
	w.updateJobStatus(job.ID, job.AssetID, job.UserID, "processing", "")

	// Processar arquivo
	err := w.processFile(job)
	if err != nil {
		log.Printf("Worker %d: Error processing job %s: %v", w.ID, job.ID, err)
		w.updateJobStatus(job.ID, job.AssetID, job.UserID, "failed", err.Error())
		return
	}

	// Sucesso
	w.updateJobStatus(job.ID, job.AssetID, job.UserID, "completed", "")
	log.Printf("Worker %d: Job %s completed successfully", w.ID, job.ID)
}

func (w *Worker) processFile(job *queue.ProcessingJob) error {
	// Criar diretório de cache se não existir
	cacheDir := "cache"
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return fmt.Errorf("erro ao criar diretório de cache: %v", err)
	}

	// Gerar caminho do cache
	log.Printf("Gerando asset path %s\n", job.AssetPath)
	filename := filepath.Base(job.AssetPath)
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%s_%s", job.UserID, filename))

	//pra fins de debug, verifica quais os arquivos existem dentro de temp/
	files, errr := os.ReadDir(cacheDir)
	if errr != nil {
		return fmt.Errorf("erro ao ler diretório de cache: %v", errr)
	}

	for _, file := range files {
		log.Printf("Arquivo no cache: %s", file.Name())
	}
	/// debug

	// Aplicar watermark baseado no tipo
	watermarkText := fmt.Sprintf("%s (%s)", job.UserID, job.UserEmail)
	ext := filepath.Ext(job.AssetPath)

	var err error
	switch ext {
	case ".pdf":
		err = watermarker.AddPDFWatermark(job.AssetPath, cachePath, watermarkText)
	case ".mp4", ".mov":
		// Verificar tamanho do arquivo para escolher estratégia
		fileInfo, statErr := os.Stat(job.AssetPath)
		if statErr != nil {
			return fmt.Errorf("erro ao verificar arquivo: %v", statErr)
		}

		// Arquivos maiores que 500MB usam processamento ultra-rápido
		if fileInfo.Size() > 500*1024*1024 {
			log.Printf("Arquivo grande detectado (%.2f MB), usando processamento otimizado",
				float64(fileInfo.Size())/(1024*1024))
			err = watermarker.AddVideoWatermarkLarge(job.AssetPath, cachePath, watermarkText)
		} else {
			err = watermarker.AddVideoWatermark(job.AssetPath, cachePath, watermarkText)
		}
	default:
		return fmt.Errorf("tipo de arquivo não suportado: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("erro ao aplicar watermark: %v", err)
	}

	// Atualizar cache path no banco
	var processedAsset models.ProcessedAsset
	if err := database.DB.Where("asset_id = ? AND user_id = ?", job.AssetID, job.UserID).First(&processedAsset).Error; err != nil {
		return fmt.Errorf("erro ao encontrar processed asset: %v", err)
	}

	now := time.Now()
	processedAsset.CachePath = cachePath
	processedAsset.ProcessedAt = &now

	return database.DB.Save(&processedAsset).Error
}

func (w *Worker) updateJobStatus(jobID, assetID, userID, status, errorMsg string) {
	// Atualizar status no Redis
	w.queue.SetJobStatus(jobID, status)

	// Atualizar status no banco
	var processedAsset models.ProcessedAsset
	if err := database.DB.Where("asset_id = ? AND user_id = ?", assetID, userID).First(&processedAsset).Error; err != nil {
		log.Printf("Error finding processed asset: %v", err)
		return
	}

	processedAsset.Status = status
	processedAsset.ErrorMsg = errorMsg

	if err := database.DB.Save(&processedAsset).Error; err != nil {
		log.Printf("Error updating processed asset status: %v", err)
	}
}
