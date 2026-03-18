// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package catalog

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"vega-backend/interfaces"
	mock_interfaces "vega-backend/interfaces/mock"
)

// mockCipher 实现 kwcrypto.Cipher 接口用于测试
// 注：kwcrypto.Cipher 是外部库接口，无 mockgen 生成的版本，手写 mock 是合理的
type mockCipher struct {
	decryptFunc func(ciphertext string) (string, error)
}

func (m *mockCipher) Encrypt(plaintext string) (string, error) {
	return "encrypted_" + plaintext, nil
}

func (m *mockCipher) Decrypt(ciphertext string) (string, error) {
	return m.decryptFunc(ciphertext)
}

func (m *mockCipher) Signature(data string) (string, error) {
	return "", nil
}

// ===== validateAndDecryptSensitiveFields =====

func TestValidateAndDecrypt_NoCipher(t *testing.T) {
	cs := &catalogService{cipher: nil}
	config := map[string]any{"password": "secret123", "host": "localhost"}

	decrypted, err := cs.validateAndDecryptSensitiveFields([]string{"password"}, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decrypted["password"] != "secret123" {
		t.Errorf("expected 'secret123', got '%v'", decrypted["password"])
	}
	if config["password"] != "secret123" {
		t.Errorf("original config should not be modified")
	}
}

func TestValidateAndDecrypt_WithCipher_Success(t *testing.T) {
	cs := &catalogService{
		cipher: &mockCipher{
			decryptFunc: func(ciphertext string) (string, error) {
				return "decrypted_" + ciphertext, nil
			},
		},
	}
	config := map[string]any{"password": "rsa_ciphertext", "host": "localhost"}

	decrypted, err := cs.validateAndDecryptSensitiveFields([]string{"password"}, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decrypted["password"] != "decrypted_rsa_ciphertext" {
		t.Errorf("expected 'decrypted_rsa_ciphertext', got '%v'", decrypted["password"])
	}
	if config["password"] != EncryptedPrefix+"rsa_ciphertext" {
		t.Errorf("expected ENC: prefix, got '%v'", config["password"])
	}
	if decrypted["host"] != "localhost" {
		t.Errorf("non-sensitive field should be unchanged")
	}
}

func TestValidateAndDecrypt_WithCipher_DecryptFails(t *testing.T) {
	cs := &catalogService{
		cipher: &mockCipher{
			decryptFunc: func(ciphertext string) (string, error) {
				return "", fmt.Errorf("invalid ciphertext")
			},
		},
	}
	config := map[string]any{"password": "bad_data"}

	_, err := cs.validateAndDecryptSensitiveFields([]string{"password"}, config)
	if err == nil {
		t.Fatal("expected error for invalid ciphertext")
	}
	if !strings.Contains(err.Error(), "password") {
		t.Errorf("error should mention field name, got: %v", err)
	}
}

func TestValidateAndDecrypt_EmptyValue(t *testing.T) {
	cs := &catalogService{
		cipher: &mockCipher{
			decryptFunc: func(ciphertext string) (string, error) {
				t.Fatal("should not be called for empty value")
				return "", nil
			},
		},
	}
	config := map[string]any{"password": ""}

	_, err := cs.validateAndDecryptSensitiveFields([]string{"password"}, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAndDecrypt_NonStringValue(t *testing.T) {
	cs := &catalogService{
		cipher: &mockCipher{
			decryptFunc: func(ciphertext string) (string, error) {
				t.Fatal("should not be called for non-string value")
				return "", nil
			},
		},
	}
	config := map[string]any{"password": 12345}

	_, err := cs.validateAndDecryptSensitiveFields([]string{"password"}, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ===== decryptSensitiveFields =====

func TestDecrypt_NoCipher(t *testing.T) {
	cs := &catalogService{cipher: nil}
	config := map[string]any{"password": "ENC:ciphertext"}

	decrypted, err := cs.decryptSensitiveFields([]string{"password"}, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decrypted["password"] != "ENC:ciphertext" {
		t.Errorf("expected original value, got '%v'", decrypted["password"])
	}
}

func TestDecrypt_WithCipher_Success(t *testing.T) {
	cs := &catalogService{
		cipher: &mockCipher{
			decryptFunc: func(ciphertext string) (string, error) {
				return "plaintext_" + ciphertext, nil
			},
		},
	}
	config := map[string]any{"password": "ENC:rsa_data"}

	decrypted, err := cs.decryptSensitiveFields([]string{"password"}, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decrypted["password"] != "plaintext_rsa_data" {
		t.Errorf("expected 'plaintext_rsa_data', got '%v'", decrypted["password"])
	}
}

func TestDecrypt_MissingEncPrefix(t *testing.T) {
	cs := &catalogService{
		cipher: &mockCipher{
			decryptFunc: func(ciphertext string) (string, error) {
				return "", nil
			},
		},
	}
	config := map[string]any{"password": "no_prefix_value"}

	_, err := cs.decryptSensitiveFields([]string{"password"}, config)
	if err == nil {
		t.Fatal("expected error for missing ENC: prefix")
	}
	if !strings.Contains(err.Error(), "not encrypted") {
		t.Errorf("expected 'not encrypted' error, got: %v", err)
	}
}

func TestDecrypt_DecryptFails(t *testing.T) {
	cs := &catalogService{
		cipher: &mockCipher{
			decryptFunc: func(ciphertext string) (string, error) {
				return "", fmt.Errorf("corrupted data")
			},
		},
	}
	config := map[string]any{"password": "ENC:bad_data"}

	_, err := cs.decryptSensitiveFields([]string{"password"}, config)
	if err == nil {
		t.Fatal("expected error for corrupted data")
	}
	if !strings.Contains(err.Error(), "password") {
		t.Errorf("error should mention field name, got: %v", err)
	}
}

// ===== CheckExistByID（使用 mockgen 生成的 mock） =====

func TestCheckExistByID_Found(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
	mockCA.EXPECT().GetByID(gomock.Any(), "test-id").
		Return(&interfaces.Catalog{ID: "test-id"}, nil)

	cs := &catalogService{ca: mockCA}
	exists, err := cs.CheckExistByID(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected catalog to exist")
	}
}

func TestCheckExistByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
	mockCA.EXPECT().GetByID(gomock.Any(), "missing-id").
		Return(nil, nil)

	cs := &catalogService{ca: mockCA}
	exists, err := cs.CheckExistByID(context.Background(), "missing-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected catalog to not exist")
	}
}

func TestCheckExistByID_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
	mockCA.EXPECT().GetByID(gomock.Any(), "test-id").
		Return(nil, fmt.Errorf("db error"))

	cs := &catalogService{ca: mockCA}
	_, err := cs.CheckExistByID(context.Background(), "test-id")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== CheckExistByName =====

func TestCheckExistByName_Found(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCA := mock_interfaces.NewMockCatalogAccess(ctrl)
	mockCA.EXPECT().GetByName(gomock.Any(), "test").
		Return(&interfaces.Catalog{Name: "test"}, nil)

	cs := &catalogService{ca: mockCA}
	exists, err := cs.CheckExistByName(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected catalog to exist")
	}
}

// ===== TestConnection =====

func TestTestConnection_NilCatalog(t *testing.T) {
	cs := &catalogService{}
	_, err := cs.TestConnection(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil catalog")
	}
}

func TestTestConnection_Valid(t *testing.T) {
	cs := &catalogService{}
	catalog := &interfaces.Catalog{
		CatalogHealthCheckStatus: interfaces.CatalogHealthCheckStatus{
			HealthCheckStatus: interfaces.CatalogHealthStatusHealthy,
			LastCheckTime:     1234567890,
		},
	}
	result, err := cs.TestConnection(context.Background(), catalog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HealthCheckStatus != interfaces.CatalogHealthStatusHealthy {
		t.Errorf("expected healthy status, got %s", result.HealthCheckStatus)
	}
}

// ===== DeleteByIDs empty =====

func TestDeleteByIDs_Empty(t *testing.T) {
	cs := &catalogService{}
	err := cs.DeleteByIDs(context.Background(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
