package watermarker

import (
	"fmt"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func AddPDFWatermark(inputPath, outputPath, userID string) error {
	description := fmt.Sprintf(`Licensed to: %s`, userID)

	wm, err := pdfcpu.ParseTextWatermarkDetails(description, "", false, types.POINTS)
	if err != nil {
		return fmt.Errorf("erro ao criar configuração de marca d'água: %v", err)
	}

	if err := api.AddWatermarksFile(inputPath, outputPath, nil, wm, nil); err != nil {
		return fmt.Errorf("erro ao adicionar marca d'água: %v", err)
	}

	return nil
}
