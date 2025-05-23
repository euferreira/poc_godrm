package queue

import (
	"encoding/json"
	"fmt"
	"projeto_drm/poc/internal/watermarker"
	"time"
)

func StartWorker() {
	go func() {
		for {
			result, err := RedisClient.BRPop(Ctx, 0*time.Second, queueName).Result()
			if err != nil {
				fmt.Println("Erro ao ler da fila:", err)
				continue
			}

			var job AssetJob
			if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
				fmt.Println("Erro ao decodificar job:", err)
				continue
			}

			fmt.Println("Processando asset:", job.ID)

			switch job.Type {
			case "application/pdf":
				watermarker.AddPDFWatermark(job.Path, job.Path, fmt.Sprintf("User %d", job.ID))
			case "video/mp4", "video/quicktime":
				watermarker.AddVideoWatermark(job.Path, job.Path, fmt.Sprintf("User %d", job.ID))
			default:
				fmt.Println("Tipo de arquivo não suportado para watermark:", job.Type)
			}
		}
	}()
}
