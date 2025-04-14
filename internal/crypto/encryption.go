package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

var secretKey = []byte("sua-chave-secreta-32-bytes-aqui!")

func EncryptWithExpiration(content []byte, expirationTime time.Time) ([]byte, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	expBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(expBytes, uint64(expirationTime.Unix()))

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	encrypted := make([]byte, len(content))
	stream.XORKeyStream(encrypted, content)

	result := make([]byte, len(iv)+len(expBytes)+len(encrypted))
	copy(result, iv)
	copy(result[len(iv):], expBytes)
	copy(result[len(iv)+len(expBytes):], encrypted)

	return result, nil
}

func DecryptWithExpiration(encryptedData []byte) ([]byte, time.Time, error) {
	// Verificar se os dados são suficientemente longos
	if len(encryptedData) < aes.BlockSize+8 {
		return nil, time.Time{}, fmt.Errorf("dados criptografados inválidos")
	}

	// Extrair IV, tempo de expiração e dados criptografados
	iv := encryptedData[:aes.BlockSize]
	expBytes := encryptedData[aes.BlockSize : aes.BlockSize+8]
	encryptedContent := encryptedData[aes.BlockSize+8:]

	// Obter o tempo de expiração
	expirationTime := time.Unix(int64(binary.BigEndian.Uint64(expBytes)), 0)

	// Criar o bloco AES
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("erro ao criar cipher AES: %v", err)
	}

	// Criar o decodificador CFB
	stream := cipher.NewCFBDecrypter(block, iv)

	// Descriptografar o conteúdo
	stream.XORKeyStream(encryptedContent, encryptedContent)

	return encryptedContent, expirationTime, nil
}
