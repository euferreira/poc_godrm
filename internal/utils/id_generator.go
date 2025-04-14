package utils

import (
	"github.com/google/uuid"
	"strings"
)

func GenerateAssetID() string {
	id := uuid.New().String()
	return strings.ReplaceAll(id, "-", "")
}
