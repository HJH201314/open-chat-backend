package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"sync"
)

type EncryptService struct {
	PublicKey    *rsa.PublicKey
	PublicKeyPEM string
	privateKey   *rsa.PrivateKey
}

func (s *EncryptService) Decrypt(ciphertext string) (string, error) {
	// Decode base64 encrypted data
	encryptedBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", errors.New("failed to decode base64")
	}

	// Decrypt with private key OAEP
	decryptedBytes, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, s.privateKey, encryptedBytes, nil)
	if err != nil {
		return "", errors.New("failed to decrypt data")
	}

	return string(decryptedBytes), nil
}

var (
	encryptServiceInstance *EncryptService
	encryptServiceOnce     sync.Once
)

func InitEncryptService() {
	encryptServiceOnce.Do(
		func() {
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				panic(err)
			}
			publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
			publicKeyPEM := pem.EncodeToMemory(
				&pem.Block{
					Type:  "PUBLIC KEY",
					Bytes: publicKeyBytes,
				},
			)
			encryptServiceInstance = &EncryptService{
				privateKey:   privateKey,
				PublicKey:    &privateKey.PublicKey,
				PublicKeyPEM: string(publicKeyPEM),
			}
		},
	)
}

func GetEncryptService() *EncryptService {
	if encryptServiceInstance == nil {
		InitEncryptService()
	}
	return encryptServiceInstance
}
