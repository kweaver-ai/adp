// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/mock/gomock"

	"vega-backend/interfaces"
	mock_interfaces "vega-backend/interfaces/mock"
)

// newTestService 使用 mockgen 生成的 mock 构建 resourceService
func newTestService(t *testing.T) (*resourceService, *mock_interfaces.MockResourceAccess, *mock_interfaces.MockPermissionService, *mock_interfaces.MockDatasetService, *mock_interfaces.MockUserMgmtService) {
	ctrl := gomock.NewController(t)
	mockRA := mock_interfaces.NewMockResourceAccess(ctrl)
	mockPS := mock_interfaces.NewMockPermissionService(ctrl)
	mockDS := mock_interfaces.NewMockDatasetService(ctrl)
	mockUMS := mock_interfaces.NewMockUserMgmtService(ctrl)

	rs := &resourceService{
		ra:  mockRA,
		ps:  mockPS,
		ds:  mockDS,
		ums: mockUMS,
	}
	return rs, mockRA, mockPS, mockDS, mockUMS
}

// ===== CheckExistByID =====

func TestCheckExistByID_Found(t *testing.T) {
	rs, mockRA, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByID(gomock.Any(), "r1").
		Return(&interfaces.Resource{ID: "r1"}, nil)

	exists, err := rs.CheckExistByID(context.Background(), "r1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected resource to exist")
	}
}

func TestCheckExistByID_NotFound(t *testing.T) {
	rs, mockRA, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByID(gomock.Any(), "missing").
		Return(nil, nil)

	exists, err := rs.CheckExistByID(context.Background(), "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected resource to not exist")
	}
}

func TestCheckExistByID_Error(t *testing.T) {
	rs, mockRA, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByID(gomock.Any(), "r1").
		Return(nil, fmt.Errorf("db error"))

	_, err := rs.CheckExistByID(context.Background(), "r1")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== CheckExistByName =====

func TestCheckExistByName_Found(t *testing.T) {
	rs, mockRA, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByName(gomock.Any(), "cat1", "test").
		Return(&interfaces.Resource{Name: "test"}, nil)

	exists, err := rs.CheckExistByName(context.Background(), "cat1", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected resource to exist")
	}
}

func TestCheckExistByName_NotFound(t *testing.T) {
	rs, mockRA, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByName(gomock.Any(), "cat1", "missing").
		Return(nil, nil)

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
	rs, mockRA, mockPS, _, mockUMS := newTestService(t)
	mockRA.EXPECT().GetByID(gomock.Any(), "r1").
		Return(&interfaces.Resource{ID: "r1", Name: "test"}, nil)
	mockPS.EXPECT().FilterResources(gomock.Any(), interfaces.RESOURCE_TYPE_RESOURCE, []string{"r1"}, gomock.Any(), true).
		Return(map[string]interfaces.PermissionResourceOps{
			"r1": {ResourceID: "r1", Operations: []string{"view_detail"}},
		}, nil)
	mockUMS.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

	resource, err := rs.GetByID(context.Background(), "r1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resource.ID != "r1" {
		t.Errorf("expected ID 'r1', got '%s'", resource.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	rs, mockRA, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByID(gomock.Any(), "missing").
		Return(nil, nil)

	_, err := rs.GetByID(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for not found resource")
	}
}

func TestGetByID_DBError(t *testing.T) {
	rs, mockRA, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByID(gomock.Any(), "r1").
		Return(nil, fmt.Errorf("db error"))

	_, err := rs.GetByID(context.Background(), "r1")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== GetByIDs =====

func TestGetByIDs_Success(t *testing.T) {
	rs, mockRA, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByIDs(gomock.Any(), []string{"r1", "r2"}).
		Return([]*interfaces.Resource{{ID: "r1"}, {ID: "r2"}}, nil)

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
	rs, mockRA, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByCatalogID(gomock.Any(), "cat1").
		Return([]*interfaces.Resource{{ID: "r1", CatalogID: "cat1"}}, nil)

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
	rs, _, _, _, _ := newTestService(t)
	err := rs.DeleteByIDs(context.Background(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteByIDs_Success(t *testing.T) {
	rs, mockRA, mockPS, _, _ := newTestService(t)
	mockPS.EXPECT().FilterResources(gomock.Any(), interfaces.RESOURCE_TYPE_RESOURCE, []string{"r1"}, gomock.Any(), true).
		Return(map[string]interfaces.PermissionResourceOps{
			"r1": {ResourceID: "r1"},
		}, nil)
	mockRA.EXPECT().GetByIDs(gomock.Any(), []string{"r1"}).
		Return([]*interfaces.Resource{{ID: "r1", Category: "table"}}, nil)
	mockRA.EXPECT().DeleteByIDs(gomock.Any(), []string{"r1"}).Return(nil)
	mockPS.EXPECT().DeleteResources(gomock.Any(), interfaces.RESOURCE_TYPE_RESOURCE, []string{"r1"}).Return(nil)

	err := rs.DeleteByIDs(context.Background(), []string{"r1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ===== Create =====

func TestCreate_Success(t *testing.T) {
	rs, mockRA, mockPS, _, _ := newTestService(t)
	mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockRA.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	mockPS.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

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
	rs, mockRA, mockPS, _, _ := newTestService(t)
	mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockRA.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	mockPS.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

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
	rs, mockRA, mockPS, _, _ := newTestService(t)
	mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockRA.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("db error"))

	_, err := rs.Create(context.Background(), &interfaces.ResourceRequest{
		Name: "test-resource",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== UpdateStatus =====

func TestUpdateStatus_Success(t *testing.T) {
	rs, mockRA, _, _, _ := newTestService(t)
	mockRA.EXPECT().UpdateStatus(gomock.Any(), "r1", "active", "").Return(nil)

	err := rs.UpdateStatus(context.Background(), "r1", "active", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateStatus_Error(t *testing.T) {
	rs, mockRA, _, _, _ := newTestService(t)
	mockRA.EXPECT().UpdateStatus(gomock.Any(), "r1", "active", "").
		Return(fmt.Errorf("db error"))

	err := rs.UpdateStatus(context.Background(), "r1", "active", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== Update =====

func TestUpdate_NilOrigin(t *testing.T) {
	rs, _, _, _, _ := newTestService(t)
	err := rs.Update(context.Background(), "r1", &interfaces.ResourceRequest{
		OriginResource: nil,
	})
	if err == nil {
		t.Fatal("expected error for nil origin resource")
	}
}

func TestUpdate_Success(t *testing.T) {
	rs, mockRA, mockPS, _, _ := newTestService(t)
	mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockRA.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	err := rs.Update(context.Background(), "r1", &interfaces.ResourceRequest{
		OriginResource: &interfaces.Resource{ID: "r1"},
		Name:           "updated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
