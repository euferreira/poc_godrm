package watermarker

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func getCPUCount() int {
	// Usar 75% dos CPUs disponíveis para não sobrecarregar o sistema
	cpus := runtime.NumCPU()
	if cpus > 1 {
		return int(float64(cpus) * 0.75)
	}
	return 1
}

func AddVideoWatermark(inputPath, outputPath, userID string) error {
	cpuCount := getCPUCount()

	cmd := exec.Command("ffmpeg",
		// Input
		"-i", inputPath,

		// Hardware acceleration (se disponível)
		"-hwaccel", "auto",

		// Video filters - watermark otimizado
		"-vf", fmt.Sprintf(
			"drawtext=text='Licensed to: %s':x=w-tw-20:y=h-th-20:fontsize=32:fontcolor=white@0.8:borderw=2:bordercolor=black@0.8:box=1:boxcolor=black@0.3:boxborderw=5",
			userID,
		),

		// Encoding otimizado
		"-c:v", "libx264",
		"-preset", "faster", // Mais rápido que "fast"
		"-crf", "28", // Qualidade um pouco menor, muito mais rápido
		"-profile:v", "high", // Profile otimizado
		"-level", "4.0", // Level otimizado

		// Multi-threading otimizado
		"-threads", fmt.Sprintf("%d", cpuCount),
		"-thread_type", "slice+frame",

		// Audio - copy sem reencoding
		"-c:a", "copy",

		// Otimizações gerais
		"-movflags", "+faststart", // Otimiza para streaming
		"-avoid_negative_ts", "make_zero",

		// Output
		"-y", // Sobrescrever arquivo se existir
		outputPath,
	)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	fmt.Printf("Processando vídeo com %d threads...\n", cpuCount)
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

func AddVideoWatermarkLarge(inputPath, outputPath, userID string) error {
	cpuCount := getCPUCount()

	cmd := exec.Command("ffmpeg",
		"-i", inputPath,

		// Hardware acceleration
		"-hwaccel", "auto",

		// Watermark simplificado para arquivos grandes
		"-vf", fmt.Sprintf(
			"drawtext=text='%s':x=w-tw-10:y=10:fontsize=24:fontcolor=white@0.7:borderw=1:bordercolor=black",
			userID,
		),

		// Encoding ultra-rápido
		"-c:v", "libx264",
		"-preset", "ultrafast", // Mais rápido possível
		"-crf", "30", // Qualidade menor
		"-tune", "fastdecode", // Otimizar para decodificação rápida

		// Threading agressivo
		"-threads", fmt.Sprintf("%d", cpuCount),
		"-thread_type", "slice+frame",

		// Audio copy
		"-c:a", "copy",

		// Otimizações para arquivos grandes
		"-movflags", "+faststart",
		"-fflags", "+genpts",

		"-y",
		outputPath,
	)

	cmd.Stderr = os.Stderr

	fmt.Printf("Processando arquivo grande com configurações ultra-rápidas...\n")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao processar arquivo grande: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("arquivo de saída não foi gerado")
	}

	return nil
}
