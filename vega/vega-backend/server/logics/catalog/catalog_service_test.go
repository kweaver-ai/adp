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

	"vega-backend/interfaces"
)

// mockCipher 实现 kwcrypto.Cipher 接口用于测试
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

// mockCatalogAccess 实现 interfaces.CatalogAccess
type mockCatalogAccess struct {
	getByIDResult  *interfaces.Catalog
	getByIDErr     error
	getByNameResult *interfaces.Catalog
	getByNameErr   error
}

func (m *mockCatalogAccess) Create(ctx context.Context, catalog *interfaces.Catalog) error { return nil }
func (m *mockCatalogAccess) GetByID(ctx context.Context, id string) (*interfaces.Catalog, error) {
	return m.getByIDResult, m.getByIDErr
}
func (m *mockCatalogAccess) GetByIDs(ctx context.Context, ids []string) ([]*interfaces.Catalog, error) {
	return nil, nil
}
func (m *mockCatalogAccess) GetByName(ctx context.Context, name string) (*interfaces.Catalog, error) {
	return m.getByNameResult, m.getByNameErr
}
func (m *mockCatalogAccess) List(ctx context.Context, params interfaces.CatalogsQueryParams) ([]*interfaces.Catalog, int64, error) {
	return nil, 0, nil
}
func (m *mockCatalogAccess) Update(ctx context.Context, catalog *interfaces.Catalog) error { return nil }
func (m *mockCatalogAccess) DeleteByIDs(ctx context.Context, ids []string) error           { return nil }
func (m *mockCatalogAccess) UpdateMetadata(ctx context.Context, id string, metadata map[string]any) error {
	return nil
}
func (m *mockCatalogAccess) UpdateHealthCheckStatus(ctx context.Context, id string, status interfaces.CatalogHealthCheckStatus) error {
	return nil
}

// ===== validateAndDecryptSensitiveFields =====

func TestValidateAndDecrypt_NoCipher(t *testing.T) {
	cs := &catalogService{cipher: nil}
	config := map[string]any{"password": "secret123", "host": "localhost"}

	decrypted, err := cs.validateAndDecryptSensitiveFields([]string{"password"}, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// cipher 为 nil 时，直接返回拷贝
	if decrypted["password"] != "secret123" {
		t.Errorf("expected 'secret123', got '%v'", decrypted["password"])
	}
	// 原始 config 不应被修改
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
	// 解密后的明文
	if decrypted["password"] != "decrypted_rsa_ciphertext" {
		t.Errorf("expected 'decrypted_rsa_ciphertext', got '%v'", decrypted["password"])
	}
	// 原始 config 应加上 ENC: 前缀
	if config["password"] != EncryptedPrefix+"rsa_ciphertext" {
		t.Errorf("expected ENC: prefix, got '%v'", config["password"])
	}
	// 非敏感字段不变
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
	// cipher 为 nil 时返回原值拷贝
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

// ===== CheckExistByID =====

func TestCheckExistByID_Found(t *testing.T) {
	cs := &catalogService{
		ca: &mockCatalogAccess{
			getByIDResult: &interfaces.Catalog{ID: "test-id"},
		},
	}
	exists, err := cs.CheckExistByID(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected catalog to exist")
	}
}

func TestCheckExistByID_NotFound(t *testing.T) {
	cs := &catalogService{
		ca: &mockCatalogAccess{
			getByIDResult: nil,
		},
	}
	exists, err := cs.CheckExistByID(context.Background(), "missing-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected catalog to not exist")
	}
}

func TestCheckExistByID_Error(t *testing.T) {
	cs := &catalogService{
		ca: &mockCatalogAccess{
			getByIDErr: fmt.Errorf("db error"),
		},
	}
	_, err := cs.CheckExistByID(context.Background(), "test-id")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== CheckExistByName =====

func TestCheckExistByName_Found(t *testing.T) {
	cs := &catalogService{
		ca: &mockCatalogAccess{
			getByNameResult: &interfaces.Catalog{Name: "test"},
		},
	}
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
