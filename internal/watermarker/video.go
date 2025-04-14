package watermarker

import (
	"fmt"
	"os"
	"os/exec"
)

func AddVideoWatermark(inputPath, outputPath, userID string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-vf", fmt.Sprintf(
			"drawtext=text='Licensed to: %s':x=if(gte(t\\,0)*lt(t\\,10)\\,(w-text_w)/2\\,if(gte(t\\,10)*lt(t\\,20)\\,0\\,(w-text_w))):y=if(gte(t\\,0)*lt(t\\,10)\\,(h-text_h)/2\\,if(gte(t\\,10)*lt(t\\,20)\\,0\\,(h-text_h))):fontsize=80:fontcolor=white@0.8:borderw=4:bordercolor=black:enable='between(t,0,30)'",
			userID,
		),
		"-codec:v", "libx264", "-crf", "23", "-preset", "fast",
		"-codec:a", "copy",
		outputPath,
	)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	fmt.Println("Executando comando ffmpeg:", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao adicionar watermark no vídeo: %v", err)
	}

	// Verificar se o arquivo de saída foi gerado
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("arquivo de saída não foi gerado: %v", err)
	}

	fmt.Println("Arquivo de saída gerado com sucesso:", outputPath)

	return nil
}
