package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type RedisQueue struct {
	client *redis.Client
	ctx    context.Context
}

type ProcessingJob struct {
	ID        string    `json:"id"`
	AssetID   string    `json:"asset_id"`
	UserID    string    `json:"user_id"`
	AssetPath string    `json:"asset_path"`
	AssetType string    `json:"asset_type"`
	UserEmail string    `json:"user_email"`
	CreatedAt time.Time `json:"created_at"`
}

func NewRedisQueue(redisAddr string) *RedisQueue {
	log.Println("Connecting to Redis at", redisAddr)

	return &RedisQueue{
		client: RedisClient,
		ctx:    Ctx,
	}
}

func (rq *RedisQueue) EnqueueJob(job ProcessingJob) error {
	jobJSON, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return rq.client.LPush(rq.ctx, "processing_queue", jobJSON).Err()
}

func (rq *RedisQueue) DequeueJob() (*ProcessingJob, error) {
	result, err := rq.client.BRPop(rq.ctx, 0, "processing_queue").Result()
	if err != nil {
		return nil, err
	}

	var job ProcessingJob
	err = json.Unmarshal([]byte(result[1]), &job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (rq *RedisQueue) SetJobStatus(jobID, status string) error {
	key := fmt.Sprintf("job_status:%s", jobID)
	return rq.client.Set(rq.ctx, key, status, 24*time.Hour).Err()
}

func (rq *RedisQueue) GetJobStatus(jobID string) (string, error) {
	key := fmt.Sprintf("job_status:%s", jobID)
	return rq.client.Get(rq.ctx, key).Result()
}
