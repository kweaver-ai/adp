// Code generated manually for skill tests.
package mocks

import (
	context "context"
	sql "database/sql"
	reflect "reflect"

	ormhelper "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	model "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	gomock "go.uber.org/mock/gomock"
)

type MockISkillRepository struct {
	ctrl     *gomock.Controller
	recorder *MockISkillRepositoryMockRecorder
}

type MockISkillRepositoryMockRecorder struct{ mock *MockISkillRepository }

func NewMockISkillRepository(ctrl *gomock.Controller) *MockISkillRepository {
	mock := &MockISkillRepository{ctrl: ctrl}
	mock.recorder = &MockISkillRepositoryMockRecorder{mock}
	return mock
}

func (m *MockISkillRepository) EXPECT() *MockISkillRepositoryMockRecorder { return m.recorder }

func (m *MockISkillRepository) InsertSkill(ctx context.Context, tx *sql.Tx, skill *model.SkillRepositoryDB) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertSkill", ctx, tx, skill)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockISkillRepositoryMockRecorder) InsertSkill(ctx, tx, skill any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertSkill", reflect.TypeOf((*MockISkillRepository)(nil).InsertSkill), ctx, tx, skill)
}

func (m *MockISkillRepository) UpdateSkill(ctx context.Context, tx *sql.Tx, skill *model.SkillRepositoryDB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateSkill", ctx, tx, skill)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockISkillRepositoryMockRecorder) UpdateSkill(ctx, tx, skill any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSkill", reflect.TypeOf((*MockISkillRepository)(nil).UpdateSkill), ctx, tx, skill)
}

func (m *MockISkillRepository) UpdateSkillStatus(ctx context.Context, tx *sql.Tx, skillID string, status model.SkillStatus, updateUser string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateSkillStatus", ctx, tx, skillID, status, updateUser)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockISkillRepositoryMockRecorder) UpdateSkillStatus(ctx, tx, skillID, status, updateUser any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSkillStatus", reflect.TypeOf((*MockISkillRepository)(nil).UpdateSkillStatus), ctx, tx, skillID, status, updateUser)
}

func (m *MockISkillRepository) SelectSkillByID(ctx context.Context, tx *sql.Tx, skillID string) (*model.SkillRepositoryDB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectSkillByID", ctx, tx, skillID)
	ret0, _ := ret[0].(*model.SkillRepositoryDB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockISkillRepositoryMockRecorder) SelectSkillByID(ctx, tx, skillID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectSkillByID", reflect.TypeOf((*MockISkillRepository)(nil).SelectSkillByID), ctx, tx, skillID)
}

func (m *MockISkillRepository) SelectSkillListPage(ctx context.Context, tx *sql.Tx, filter map[string]interface{}, sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) ([]*model.SkillRepositoryDB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectSkillListPage", ctx, tx, filter, sort, cursor)
	ret0, _ := ret[0].([]*model.SkillRepositoryDB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockISkillRepositoryMockRecorder) SelectSkillListPage(ctx, tx, filter, sort, cursor any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectSkillListPage", reflect.TypeOf((*MockISkillRepository)(nil).SelectSkillListPage), ctx, tx, filter, sort, cursor)
}

func (m *MockISkillRepository) CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]interface{}) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountByWhereClause", ctx, tx, filter)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockISkillRepositoryMockRecorder) CountByWhereClause(ctx, tx, filter any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountByWhereClause", reflect.TypeOf((*MockISkillRepository)(nil).CountByWhereClause), ctx, tx, filter)
}

func (m *MockISkillRepository) DeleteSkillByID(ctx context.Context, tx *sql.Tx, skillID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSkillByID", ctx, tx, skillID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockISkillRepositoryMockRecorder) DeleteSkillByID(ctx, tx, skillID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSkillByID", reflect.TypeOf((*MockISkillRepository)(nil).DeleteSkillByID), ctx, tx, skillID)
}

type MockISkillFileIndex struct {
	ctrl     *gomock.Controller
	recorder *MockISkillFileIndexMockRecorder
}

type MockISkillFileIndexMockRecorder struct{ mock *MockISkillFileIndex }

func NewMockISkillFileIndex(ctrl *gomock.Controller) *MockISkillFileIndex {
	mock := &MockISkillFileIndex{ctrl: ctrl}
	mock.recorder = &MockISkillFileIndexMockRecorder{mock}
	return mock
}

func (m *MockISkillFileIndex) EXPECT() *MockISkillFileIndexMockRecorder { return m.recorder }

func (m *MockISkillFileIndex) InsertSkillFile(ctx context.Context, tx *sql.Tx, file *model.SkillFileIndexDB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertSkillFile", ctx, tx, file)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockISkillFileIndexMockRecorder) InsertSkillFile(ctx, tx, file any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertSkillFile", reflect.TypeOf((*MockISkillFileIndex)(nil).InsertSkillFile), ctx, tx, file)
}

func (m *MockISkillFileIndex) BatchInsertSkillFiles(ctx context.Context, tx *sql.Tx, files []*model.SkillFileIndexDB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchInsertSkillFiles", ctx, tx, files)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockISkillFileIndexMockRecorder) BatchInsertSkillFiles(ctx, tx, files any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchInsertSkillFiles", reflect.TypeOf((*MockISkillFileIndex)(nil).BatchInsertSkillFiles), ctx, tx, files)
}

func (m *MockISkillFileIndex) UpdateSkillFile(ctx context.Context, tx *sql.Tx, file *model.SkillFileIndexDB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateSkillFile", ctx, tx, file)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockISkillFileIndexMockRecorder) UpdateSkillFile(ctx, tx, file any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSkillFile", reflect.TypeOf((*MockISkillFileIndex)(nil).UpdateSkillFile), ctx, tx, file)
}

func (m *MockISkillFileIndex) SelectSkillFileBySkillID(ctx context.Context, tx *sql.Tx, skillID string) ([]*model.SkillFileIndexDB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectSkillFileBySkillID", ctx, tx, skillID)
	ret0, _ := ret[0].([]*model.SkillFileIndexDB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockISkillFileIndexMockRecorder) SelectSkillFileBySkillID(ctx, tx, skillID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectSkillFileBySkillID", reflect.TypeOf((*MockISkillFileIndex)(nil).SelectSkillFileBySkillID), ctx, tx, skillID)
}

func (m *MockISkillFileIndex) SelectSkillFileByPath(ctx context.Context, tx *sql.Tx, skillID, relPath string) (*model.SkillFileIndexDB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectSkillFileByPath", ctx, tx, skillID, relPath)
	ret0, _ := ret[0].(*model.SkillFileIndexDB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockISkillFileIndexMockRecorder) SelectSkillFileByPath(ctx, tx, skillID, relPath any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectSkillFileByPath", reflect.TypeOf((*MockISkillFileIndex)(nil).SelectSkillFileByPath), ctx, tx, skillID, relPath)
}

func (m *MockISkillFileIndex) SelectSkillFileByPathHash(ctx context.Context, tx *sql.Tx, skillID, pathHash string) (*model.SkillFileIndexDB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectSkillFileByPathHash", ctx, tx, skillID, pathHash)
	ret0, _ := ret[0].(*model.SkillFileIndexDB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockISkillFileIndexMockRecorder) SelectSkillFileByPathHash(ctx, tx, skillID, pathHash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectSkillFileByPathHash", reflect.TypeOf((*MockISkillFileIndex)(nil).SelectSkillFileByPathHash), ctx, tx, skillID, pathHash)
}

func (m *MockISkillFileIndex) DeleteSkillFileBySkillID(ctx context.Context, tx *sql.Tx, skillID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSkillFileBySkillID", ctx, tx, skillID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockISkillFileIndexMockRecorder) DeleteSkillFileBySkillID(ctx, tx, skillID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSkillFileBySkillID", reflect.TypeOf((*MockISkillFileIndex)(nil).DeleteSkillFileBySkillID), ctx, tx, skillID)
}

func (m *MockISkillFileIndex) DeleteSkillFileByPath(ctx context.Context, tx *sql.Tx, skillID, relPath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSkillFileByPath", ctx, tx, skillID, relPath)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockISkillFileIndexMockRecorder) DeleteSkillFileByPath(ctx, tx, skillID, relPath any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSkillFileByPath", reflect.TypeOf((*MockISkillFileIndex)(nil).DeleteSkillFileByPath), ctx, tx, skillID, relPath)
}
