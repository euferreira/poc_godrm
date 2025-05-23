package queue

import (
	"encoding/json"
	"fmt"
)

const queueName = "asset-queue"

type AssetJob struct {
	ID   uint
	Path string
	Type string
}

func EnqueueAssetJob(job AssetJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("erro ao serializar job: %v", err)
	}

	return RedisClient.LPush(Ctx, queueName, data).Err()
}
