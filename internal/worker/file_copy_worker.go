package worker

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"projeto_drm/poc/internal/database"
	"projeto_drm/poc/internal/models"
	"projeto_drm/poc/internal/queue"
	"sync"
	"time"
)

const assetQueueName = "asset-queue"

// FileCopyWorker is responsible for copying files from temporary storage to the temp directory
type FileCopyWorker struct {
	ID   int
	quit chan bool
}

// FileCopyWorkerPool manages a pool of FileCopyWorkers
type FileCopyWorkerPool struct {
	workers []*FileCopyWorker
	wg      sync.WaitGroup
}

// NewFileCopyWorkerPool creates a new pool of FileCopyWorkers
func NewFileCopyWorkerPool(size int) *FileCopyWorkerPool {
	workers := make([]*FileCopyWorker, size)
	for i := 0; i < size; i++ {
		workers[i] = &FileCopyWorker{
			ID:   i + 1,
			quit: make(chan bool),
		}
	}

	return &FileCopyWorkerPool{
		workers: workers,
	}
}

// Start starts all workers in the pool
func (wp *FileCopyWorkerPool) Start() {
	log.Printf("Starting file copy worker pool with %d workers", len(wp.workers))

	for _, worker := range wp.workers {
		wp.wg.Add(1)
		go worker.Start(&wp.wg)
	}
}

// Stop stops all workers in the pool
func (wp *FileCopyWorkerPool) Stop() {
	log.Println("Stopping file copy worker pool...")

	for _, worker := range wp.workers {
		worker.quit <- true
	}

	wp.wg.Wait()
	log.Println("File copy worker pool stopped")
}

// Start starts the worker
func (w *FileCopyWorker) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("File copy worker %d started", w.ID)

	for {
		select {
		case <-w.quit:
			log.Printf("File copy worker %d stopping", w.ID)
			return
		default:
			// Dequeue a job from the asset queue
			result, err := queue.RedisClient.BRPop(queue.Ctx, 0, assetQueueName).Result()
			if err != nil {
				log.Printf("File copy worker %d: Error dequeueing job: %v", w.ID, err)
				time.Sleep(time.Second)
				continue
			}

			// Parse the job
			var job queue.AssetJob
			if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
				log.Printf("File copy worker %d: Error parsing job: %v", w.ID, err)
				continue
			}

			// Process the job
			w.processJob(job)
		}
	}
}

// processJob processes a file copy job
func (w *FileCopyWorker) processJob(job queue.AssetJob) {
	log.Printf("File copy worker %d: Processing job for asset %d", w.ID, job.ID)

	// Get the asset from the database
	var asset models.Asset
	if err := database.DB.First(&asset, job.ID).Error; err != nil {
		log.Printf("File copy worker %d: Error getting asset %d: %v", w.ID, job.ID, err)
		return
	}

	// Update asset status to processing
	asset.Status = models.StatusProcessing
	if err := database.DB.Save(&asset).Error; err != nil {
		log.Printf("File copy worker %d: Error updating asset status: %v", w.ID, err)
		return
	}

	// Get the temporary file path from the job
	tempFilePath := job.TempFilePath
	if tempFilePath == "" {
		log.Printf("File copy worker %d: No temporary file path provided for asset %d", w.ID, job.ID)
		updateAssetStatus(&asset, models.StatusFailed)
		return
	}

	// Create the destination file
	dstPath := job.Path
	out, err := os.Create(dstPath)
	if err != nil {
		log.Printf("File copy worker %d: Error creating destination file: %v", w.ID, err)
		updateAssetStatus(&asset, models.StatusFailed)
		return
	}
	defer out.Close()

	// Open the source file
	in, err := os.Open(tempFilePath)
	if err != nil {
		log.Printf("File copy worker %d: Error opening source file: %v", w.ID, err)
		updateAssetStatus(&asset, models.StatusFailed)
		return
	}
	defer in.Close()

	// Copy the file
	if _, err = io.Copy(out, in); err != nil {
		log.Printf("File copy worker %d: Error copying file: %v", w.ID, err)
		updateAssetStatus(&asset, models.StatusFailed)
		return
	}

	// Remove the temporary file
	if err := os.Remove(tempFilePath); err != nil {
		log.Printf("File copy worker %d: Error removing temporary file: %v", w.ID, err)
		// Continue anyway, this is not a critical error
	}

	// Update asset status to completed
	updateAssetStatus(&asset, models.StatusCompleted)
	log.Printf("File copy worker %d: Successfully copied file for asset %d", w.ID, job.ID)
}

// updateAssetStatus updates the status of an asset
func updateAssetStatus(asset *models.Asset, status string) {
	asset.Status = status
	if err := database.DB.Save(asset).Error; err != nil {
		log.Printf("Error updating asset status: %v", err)
	}
}
