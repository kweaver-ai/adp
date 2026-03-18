// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource

import (
	"context"
	"fmt"
	"testing"

	"vega-backend/interfaces"
)

// ===== Mock implementations =====

type mockResourceAccess struct {
	createErr        error
	getByIDResult    *interfaces.Resource
	getByIDErr       error
	getByIDsResult   []*interfaces.Resource
	getByIDsErr      error
	getByNameResult  *interfaces.Resource
	getByNameErr     error
	getByCatalogResult []*interfaces.Resource
	getByCatalogErr  error
	listResult       []*interfaces.Resource
	listTotal        int64
	listErr          error
	updateErr        error
	updateStatusErr  error
	deleteErr        error
}

func (m *mockResourceAccess) Create(ctx context.Context, r *interfaces.Resource) error { return m.createErr }
func (m *mockResourceAccess) GetByID(ctx context.Context, id string) (*interfaces.Resource, error) {
	return m.getByIDResult, m.getByIDErr
}
func (m *mockResourceAccess) GetByIDs(ctx context.Context, ids []string) ([]*interfaces.Resource, error) {
	return m.getByIDsResult, m.getByIDsErr
}
func (m *mockResourceAccess) GetByName(ctx context.Context, catalogID, name string) (*interfaces.Resource, error) {
	return m.getByNameResult, m.getByNameErr
}
func (m *mockResourceAccess) GetByCatalogID(ctx context.Context, catalogID string) ([]*interfaces.Resource, error) {
	return m.getByCatalogResult, m.getByCatalogErr
}
func (m *mockResourceAccess) List(ctx context.Context, params interfaces.ResourcesQueryParams) ([]*interfaces.Resource, int64, error) {
	return m.listResult, m.listTotal, m.listErr
}
func (m *mockResourceAccess) Update(ctx context.Context, r *interfaces.Resource) error { return m.updateErr }
func (m *mockResourceAccess) UpdateStatus(ctx context.Context, id, status, msg string) error {
	return m.updateStatusErr
}
func (m *mockResourceAccess) DeleteByIDs(ctx context.Context, ids []string) error { return m.deleteErr }

type mockPermissionService struct{}

func (m *mockPermissionService) CheckPermission(ctx context.Context, r interfaces.PermissionResource, ops []string) error {
	return nil
}
func (m *mockPermissionService) CreateResources(ctx context.Context, rs []interfaces.PermissionResource, ops []string) error {
	return nil
}
func (m *mockPermissionService) DeleteResources(ctx context.Context, rt string, ids []string) error {
	return nil
}
func (m *mockPermissionService) FilterResources(ctx context.Context, rt string, ids []string, ops []string, allow bool) (map[string]interfaces.PermissionResourceOps, error) {
	result := make(map[string]interfaces.PermissionResourceOps)
	for _, id := range ids {
		result[id] = interfaces.PermissionResourceOps{ResourceID: id, Operations: ops}
	}
	return result, nil
}
func (m *mockPermissionService) UpdateResource(ctx context.Context, r interfaces.PermissionResource) error {
	return nil
}

type mockUserMgmtService struct{}

func (m *mockUserMgmtService) GetAccountNames(ctx context.Context, infos []*interfaces.AccountInfo) error {
	return nil
}

type mockDatasetService struct{}

func (m *mockDatasetService) Create(ctx context.Context, r *interfaces.Resource) error  { return nil }
func (m *mockDatasetService) Update(ctx context.Context, r *interfaces.Resource) error  { return nil }
func (m *mockDatasetService) Delete(ctx context.Context, r *interfaces.Resource) error  { return nil }
func (m *mockDatasetService) ListDocuments(ctx context.Context, r *interfaces.Resource, params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	return nil, 0, nil
}
func (m *mockDatasetService) GetDocument(ctx context.Context, id string, docID string) (map[string]any, error) {
	return nil, nil
}
func (m *mockDatasetService) CreateDocuments(ctx context.Context, id string, documents []map[string]any) ([]string, error) {
	return nil, nil
}
func (m *mockDatasetService) UpdateDocument(ctx context.Context, id string, docID string, document map[string]any) error {
	return nil
}
func (m *mockDatasetService) DeleteDocument(ctx context.Context, id string, docID string) error {
	return nil
}
func (m *mockDatasetService) UpdateDocuments(ctx context.Context, id string, updateRequests []map[string]any) error {
	return nil
}
func (m *mockDatasetService) DeleteDocuments(ctx context.Context, id string, docIDs string) error {
	return nil
}
func (m *mockDatasetService) DeleteDocumentsByQuery(ctx context.Context, res *interfaces.Resource, params *interfaces.ResourceDataQueryParams) error {
	return nil
}

func newTestService(ra *mockResourceAccess) *resourceService {
	return &resourceService{
		ra:  ra,
		ps:  &mockPermissionService{},
		ums: &mockUserMgmtService{},
		ds:  &mockDatasetService{},
	}
}

// ===== CheckExistByID =====

func TestCheckExistByID_Found(t *testing.T) {
	rs := newTestService(&mockResourceAccess{
		getByIDResult: &interfaces.Resource{ID: "r1"},
	})
	exists, err := rs.CheckExistByID(context.Background(), "r1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected resource to exist")
	}
}

func TestCheckExistByID_NotFound(t *testing.T) {
	rs := newTestService(&mockResourceAccess{getByIDResult: nil})
	exists, err := rs.CheckExistByID(context.Background(), "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected resource to not exist")
	}
}

func TestCheckExistByID_Error(t *testing.T) {
	rs := newTestService(&mockResourceAccess{getByIDErr: fmt.Errorf("db error")})
	_, err := rs.CheckExistByID(context.Background(), "r1")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== CheckExistByName =====

func TestCheckExistByName_Found(t *testing.T) {
	rs := newTestService(&mockResourceAccess{
		getByNameResult: &interfaces.Resource{Name: "test"},
	})
	exists, err := rs.CheckExistByName(context.Background(), "cat1", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected resource to exist")
	}
}

func TestCheckExistByName_NotFound(t *testing.T) {
	rs := newTestService(&mockResourceAccess{getByNameResult: nil})
	exists, err := rs.CheckExistByName(context.Background(), "cat1", "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected resource to not exist")
	}
}

// ===== GetByID =====

func TestGetByID_Success(t *testing.T) {
	rs := newTestService(&mockResourceAccess{
		getByIDResult: &interfaces.Resource{ID: "r1", Name: "test"},
	})
	resource, err := rs.GetByID(context.Background(), "r1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resource.ID != "r1" {
		t.Errorf("expected ID 'r1', got '%s'", resource.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	rs := newTestService(&mockResourceAccess{getByIDResult: nil})
	_, err := rs.GetByID(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for not found resource")
	}
}

func TestGetByID_DBError(t *testing.T) {
	rs := newTestService(&mockResourceAccess{getByIDErr: fmt.Errorf("db error")})
	_, err := rs.GetByID(context.Background(), "r1")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== GetByIDs =====

func TestGetByIDs_Success(t *testing.T) {
	rs := newTestService(&mockResourceAccess{
		getByIDsResult: []*interfaces.Resource{
			{ID: "r1"}, {ID: "r2"},
		},
	})
	resources, err := rs.GetByIDs(context.Background(), []string{"r1", "r2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}
}

// ===== GetByCatalogID =====

func TestGetByCatalogID_Success(t *testing.T) {
	rs := newTestService(&mockResourceAccess{
		getByCatalogResult: []*interfaces.Resource{{ID: "r1", CatalogID: "cat1"}},
	})
	resources, err := rs.GetByCatalogID(context.Background(), "cat1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(resources))
	}
}

// ===== DeleteByIDs =====

func TestDeleteByIDs_Empty(t *testing.T) {
	rs := newTestService(&mockResourceAccess{})
	err := rs.DeleteByIDs(context.Background(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteByIDs_Success(t *testing.T) {
	rs := newTestService(&mockResourceAccess{
		getByIDsResult: []*interfaces.Resource{{ID: "r1", Category: "table"}},
	})
	err := rs.DeleteByIDs(context.Background(), []string{"r1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ===== Create =====

func TestCreate_Success(t *testing.T) {
	rs := newTestService(&mockResourceAccess{})
	id, err := rs.Create(context.Background(), &interfaces.ResourceRequest{
		Name:     "test-resource",
		Category: "table",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty ID")
	}
}

func TestCreate_WithExplicitID(t *testing.T) {
	rs := newTestService(&mockResourceAccess{})
	id, err := rs.Create(context.Background(), &interfaces.ResourceRequest{
		ID:       "custom-id",
		Name:     "test-resource",
		Category: "table",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "custom-id" {
		t.Errorf("expected 'custom-id', got '%s'", id)
	}
}

func TestCreate_DBError(t *testing.T) {
	rs := newTestService(&mockResourceAccess{createErr: fmt.Errorf("db error")})
	_, err := rs.Create(context.Background(), &interfaces.ResourceRequest{
		Name: "test-resource",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== UpdateStatus =====

func TestUpdateStatus_Success(t *testing.T) {
	rs := newTestService(&mockResourceAccess{})
	err := rs.UpdateStatus(context.Background(), "r1", "active", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateStatus_Error(t *testing.T) {
	rs := newTestService(&mockResourceAccess{updateStatusErr: fmt.Errorf("db error")})
	err := rs.UpdateStatus(context.Background(), "r1", "active", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== Update =====

func TestUpdate_NilOrigin(t *testing.T) {
	rs := newTestService(&mockResourceAccess{})
	err := rs.Update(context.Background(), "r1", &interfaces.ResourceRequest{
		OriginResource: nil,
	})
	if err == nil {
		t.Fatal("expected error for nil origin resource")
	}
}

func TestUpdate_Success(t *testing.T) {
	rs := newTestService(&mockResourceAccess{})
	err := rs.Update(context.Background(), "r1", &interfaces.ResourceRequest{
		OriginResource: &interfaces.Resource{ID: "r1"},
		Name:           "updated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
