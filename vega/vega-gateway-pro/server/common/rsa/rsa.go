package rsa

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"os"
	"strings"
)

var (
	encryptPrivateKey string
)

// InitKeys 从文件加载密钥
func InitKeys(privateKeyPath string) error {
	workDir, _ := os.Getwd()
	privateKeyBytes, err := os.ReadFile(workDir + privateKeyPath)
	if err != nil {
		return fmt.Errorf("read private key file failed: %w", err)
	}
	encryptPrivateKey = string(privateKeyBytes)

	return nil
}

func cleanPemKey(pemKey string) string {
	pemKey = strings.ReplaceAll(pemKey, "-----BEGIN PRIVATE KEY-----", "")
	pemKey = strings.ReplaceAll(pemKey, "-----END PRIVATE KEY-----", "")
	pemKey = strings.ReplaceAll(pemKey, " ", "")
	pemKey = strings.ReplaceAll(pemKey, "\n", "")
	return pemKey
}

// Decrypt RSA解密
func Decrypt(encryptedData string) (string, error) {
	cleanPrivateKey := cleanPemKey(encryptPrivateKey)
	keyBytes, err := base64.StdEncoding.DecodeString(cleanPrivateKey)
	if err != nil {
		logger.Errorf("RSA decrypt private key decode failed: %v", err)
		return "", fmt.Errorf("RSA decrypt private key decode failed: %w", err)
	}

	privKey, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		logger.Errorf("RSA decrypt private key parse failed: %v", err)
		return "", fmt.Errorf("RSA decrypt private key parse failed: %w", err)
	}

	rsaPrivKey := privKey.(*rsa.PrivateKey)
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		logger.Errorf("RSA decrypt ciphertext decode failed: %v", err)
		return "", err
	}

	plaintext, err := rsa.DecryptPKCS1v15(nil, rsaPrivKey, ciphertext)
	if err != nil {
		logger.Errorf("RSA decrypt ciphertext failed: %v", err)
		return "", fmt.Errorf("RSA decrypt ciphertext failed: %w", err)
	}

	return string(plaintext), nil
}
