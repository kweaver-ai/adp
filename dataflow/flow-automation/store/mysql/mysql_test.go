package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/utils/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/go-sql-driver/mysql"
)

var (
	testDSN = "root:test123@tcp(127.0.0.1:3306)/flow_automation_test?parseTime=true&charset=utf8mb4"
	testStore *Store
)

func TestMain(m *testing.M) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn != "" {
		testDSN = dsn
	}

	var err error
	testStore = NewStore(&StoreOption{
		DSN:      testDSN,
		MaxConns: 10,
		MinConns: 2,
	})

	err = testStore.Init()
	if err != nil {
		fmt.Printf("Failed to initialize test store: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	testStore.Close()
	os.Exit(code)
}

func cleanupTables(t *testing.T) {
	tables := []string{
		"dags", "dag_instances", "task_instances", "tokens",
		"inboxes", "outboxes", "clients", "switches", "logs", "dag_versions",
	}
	for _, table := range tables {
		_, err := testStore.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: failed to cleanup table %s: %v", table, err)
		}
	}
}

func TestStore_Init(t *testing.T) {
	store := NewStore(&StoreOption{
		DSN:      testDSN,
		MaxConns: 5,
		MinConns: 1,
	})

	err := store.Init()
	assert.NoError(t, err)
	assert.NotNil(t, store.db)

	store.Close()
}

func TestStore_Close(t *testing.T) {
	store := NewStore(&StoreOption{
		DSN:      testDSN,
		MaxConns: 5,
		MinConns: 1,
	})

	err := store.Init()
	require.NoError(t, err)

	store.Close()
}

func TestStore_WithTransaction_Commit(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()
	err := testStore.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)
		_, err := txConn.Exec("INSERT INTO switches (id, name, status, created_at, updated_at) VALUES (?, 'test_switch', true, NOW(), NOW())", "test-tx-1")
		return err
	})
	assert.NoError(t, err)

	var status bool
	err = testStore.db.QueryRow("SELECT status FROM switches WHERE name = 'test_switch'").Scan(&status)
	assert.NoError(t, err)
	assert.True(t, status)
}

func TestStore_WithTransaction_Rollback(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()
	err := testStore.WithTransaction(ctx, func(tx interface{}) error {
		txConn := tx.(*sql.Tx)
		_, err := txConn.Exec("INSERT INTO switches (id, name, status, created_at, updated_at) VALUES (?, 'test_rollback', true, NOW(), NOW())", "test-tx-rollback")
		if err != nil {
			return err
		}
		return fmt.Errorf("intentional error for rollback test")
	})
	assert.Error(t, err)

	var count int
	err = testStore.db.QueryRow("SELECT COUNT(*) FROM switches WHERE name = 'test_rollback'").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestStore_Token_CRUD(t *testing.T) {
	cleanupTables(t)

	token := &entity.Token{
		UserID:    "test-user-1",
		Token:     "test-token-value",
		ExpiresIn: 3600,
	}

	err := testStore.CreateToken(token)
	assert.NoError(t, err)
	assert.NotEmpty(t, token.ID)

	retrieved, err := testStore.GetTokenByUserID("test-user-1")
	assert.NoError(t, err)
	assert.Equal(t, "test-token-value", retrieved.Token)
	assert.Equal(t, 3600, retrieved.ExpiresIn)

	token.Token = "updated-token-value"
	token.ExpiresIn = 7200
	err = testStore.UpdateToken(token)
	assert.NoError(t, err)

	retrieved, err = testStore.GetTokenByUserID("test-user-1")
	assert.NoError(t, err)
	assert.Equal(t, "updated-token-value", retrieved.Token)
	assert.Equal(t, 7200, retrieved.ExpiresIn)

	err = testStore.DeleteToken(token.ID)
	assert.NoError(t, err)

	_, err = testStore.GetTokenByUserID("test-user-1")
	assert.NoError(t, err)
}

func TestStore_Client_CRUD(t *testing.T) {
	cleanupTables(t)

	err := testStore.CreateClient("test-client", "client-id-123", "client-secret-456")
	assert.NoError(t, err)

	client, err := testStore.GetClient("test-client")
	assert.NoError(t, err)
	assert.Equal(t, "test-client", client.ClientName)
	assert.Equal(t, "client-id-123", client.ClientID)
	assert.Equal(t, "client-secret-456", client.ClientSecret)

	err = testStore.RemoveClient("test-client")
	assert.NoError(t, err)

	_, err = testStore.GetClient("test-client")
	assert.Error(t, err)
	assert.Equal(t, data.ErrDataNotFound, err)
}

func TestStore_Dag_CRUD(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()
	dag := &entity.Dag{
		UserID: "test-user-1",
		Name:   "test-dag",
		Desc:   "test description",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks: []entity.Task{
			{ID: "task-1", Name: "Task 1"},
			{ID: "task-2", Name: "Task 2"},
		},
		Type:       "test-type",
		PolicyType: "test-policy",
		Priority:   "high",
	}

	id, err := testStore.CreateDag(ctx, dag)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	retrieved, err := testStore.GetDag(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "test-dag", retrieved.Name)
	assert.Equal(t, "test description", retrieved.Desc)
	assert.Equal(t, 2, len(retrieved.Tasks))
	assert.Equal(t, entity.DagStatusNormal, retrieved.Status)

	retrieved.Name = "updated-dag"
	retrieved.Desc = "updated description"
	err = testStore.UpdateDag(ctx, retrieved)
	assert.NoError(t, err)

	updated, err := testStore.GetDag(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "updated-dag", updated.Name)
	assert.Equal(t, "updated description", updated.Desc)

	err = testStore.BatchDeleteDagWithTransaction(ctx, []string{id})
	assert.NoError(t, err)

	_, err = testStore.GetDag(ctx, id)
	assert.Error(t, err)
}

func TestStore_Dag_BatchCreate(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()
	dags := []*entity.Dag{
		{
			UserID: "test-user-1",
			Name:   "batch-dag-1",
			Trigger: entity.Trigger{
				Type: entity.DagTriggerManual,
			},
			Status: entity.DagStatusNormal,
			Tasks:  []entity.Task{{ID: "task-1"}},
		},
		{
			UserID: "test-user-1",
			Name:   "batch-dag-2",
			Trigger: entity.Trigger{
				Type: entity.DagTriggerManual,
			},
			Status: entity.DagStatusNormal,
			Tasks:  []entity.Task{{ID: "task-2"}},
		},
	}

	result, err := testStore.BatchCreateDag(ctx, dags)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	list, err := testStore.ListDag(ctx, &mod.ListDagInput{
		UserID: "test-user-1",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 2)
}

func TestStore_Dag_GetByFields(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()
	dag := &entity.Dag{
		UserID: "test-user-fields",
		Name:   "test-dag-fields",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
		Type:   "special-type",
	}

	id, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	retrieved, err := testStore.GetDagByFields(ctx, map[string]interface{}{
		"id": id,
	})
	assert.NoError(t, err)
	assert.Equal(t, "test-dag-fields", retrieved.Name)

	retrieved, err = testStore.GetDagByFields(ctx, map[string]interface{}{
		"userid": "test-user-fields",
		"type":   "special-type",
	})
	assert.NoError(t, err)
	assert.Equal(t, "test-dag-fields", retrieved.Name)
}

func TestStore_DagInstance_CRUD(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag := &entity.Dag{
		UserID: "test-user-dagins",
		Name:   "test-dag-for-instance",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	dagID, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	dagIns := &entity.DagInstance{
		DagID:   dagID,
		Trigger: entity.Trigger{Type: entity.DagTriggerManual},
		UserID:  "test-user-dagins",
		Status:  entity.DagInstanceStatusInit,
		Vars: entity.DagInstanceVars{
			"var1": {Value: "value1"},
		},
		Priority: "high",
	}

	id, err := testStore.CreateDagIns(ctx, dagIns)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	retrieved, err := testStore.GetDagInstance(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, dagID, retrieved.DagID)
	assert.Equal(t, entity.DagInstanceStatusInit, retrieved.Status)
	assert.Equal(t, "high", retrieved.Priority)

	retrieved.Status = entity.DagInstanceStatusRunning
	err = testStore.UpdateDagIns(ctx, retrieved)
	assert.NoError(t, err)

	updated, err := testStore.GetDagInstance(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, entity.DagInstanceStatusRunning, updated.Status)

	err = testStore.BatchDeleteDagIns(ctx, []string{id})
	assert.NoError(t, err)

	_, err = testStore.GetDagInstance(ctx, id)
	assert.Error(t, err)
}

func TestStore_DagInstance_Patch(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag := &entity.Dag{
		UserID: "test-user-patch",
		Name:   "test-dag-patch",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	dagID, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	dagIns := &entity.DagInstance{
		DagID:   dagID,
		Trigger: entity.Trigger{Type: entity.DagTriggerManual},
		UserID:  "test-user-patch",
		Status:  entity.DagInstanceStatusInit,
	}
	id, err := testStore.CreateDagIns(ctx, dagIns)
	require.NoError(t, err)

	patchIns := &entity.DagInstance{
		ID:     id,
		Status: entity.DagInstanceStatusSuccess,
		Reason: "completed successfully",
	}
	err = testStore.PatchDagIns(ctx, patchIns)
	assert.NoError(t, err)

	retrieved, err := testStore.GetDagInstance(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, entity.DagInstanceStatusSuccess, retrieved.Status)
	assert.Equal(t, "completed successfully", retrieved.Reason)
}

func TestStore_TaskInstance_CRUD(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag := &entity.Dag{
		UserID: "test-user-task",
		Name:   "test-dag-task",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	dagID, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	dagIns := &entity.DagInstance{
		DagID:   dagID,
		Trigger: entity.Trigger{Type: entity.DagTriggerManual},
		UserID:  "test-user-task",
		Status:  entity.DagInstanceStatusInit,
	}
	dagInsID, err := testStore.CreateDagIns(ctx, dagIns)
	require.NoError(t, err)

	taskIns := &entity.TaskInstance{
		DagInsID: dagInsID,
		TaskID:   "task-1",
		Status:   entity.TaskInstanceStatusInit,
		Results: map[string]interface{}{
			"result1": "value1",
		},
	}

	err = testStore.CreateTaskIns(ctx, taskIns)
	assert.NoError(t, err)
	assert.NotEmpty(t, taskIns.ID)

	retrieved, err := testStore.GetTaskIns(ctx, taskIns.ID)
	assert.NoError(t, err)
	assert.Equal(t, dagInsID, retrieved.DagInsID)
	assert.Equal(t, "task-1", retrieved.TaskID)
	assert.Equal(t, entity.TaskInstanceStatusInit, retrieved.Status)

	retrieved.Status = entity.TaskInstanceStatusSuccess
	err = testStore.UpdateTaskIns(ctx, retrieved)
	assert.NoError(t, err)

	updated, err := testStore.GetTaskIns(ctx, taskIns.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity.TaskInstanceStatusSuccess, updated.Status)

	err = testStore.BatchDeleteTaskIns(ctx, []string{taskIns.ID})
	assert.NoError(t, err)

	_, err = testStore.GetTaskIns(ctx, taskIns.ID)
	assert.Error(t, err)
}

func TestStore_TaskInstance_Patch(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag := &entity.Dag{
		UserID: "test-user-task-patch",
		Name:   "test-dag-task-patch",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	dagID, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	dagIns := &entity.DagInstance{
		DagID:   dagID,
		Trigger: entity.Trigger{Type: entity.DagTriggerManual},
		UserID:  "test-user-task-patch",
		Status:  entity.DagInstanceStatusInit,
	}
	dagInsID, err := testStore.CreateDagIns(ctx, dagIns)
	require.NoError(t, err)

	taskIns := &entity.TaskInstance{
		DagInsID: dagInsID,
		TaskID:   "task-1",
		Status:   entity.TaskInstanceStatusInit,
	}
	err = testStore.CreateTaskIns(ctx, taskIns)
	require.NoError(t, err)

	patchTask := &entity.TaskInstance{
		ID:     taskIns.ID,
		Status: entity.TaskInstanceStatusRunning,
		Reason: "started execution",
	}
	err = testStore.PatchTaskIns(ctx, patchTask)
	assert.NoError(t, err)

	retrieved, err := testStore.GetTaskIns(ctx, taskIns.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity.TaskInstanceStatusRunning, retrieved.Status)
	assert.Equal(t, "started execution", retrieved.Reason)
}

func TestStore_ListDag(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		dag := &entity.Dag{
			UserID: "test-user-list",
			Name:   fmt.Sprintf("list-dag-%d", i),
			Trigger: entity.Trigger{
				Type: entity.DagTriggerManual,
			},
			Status: entity.DagStatusNormal,
			Tasks:  []entity.Task{{ID: "task-1"}},
			Type:   "list-type",
		}
		_, err := testStore.CreateDag(ctx, dag)
		require.NoError(t, err)
	}

	list, err := testStore.ListDag(ctx, &mod.ListDagInput{
		UserID: "test-user-list",
		Type:   "list-type",
		Limit:  10,
		Offset: 0,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 5)

	count, err := testStore.ListDagCount(ctx, &mod.ListDagInput{
		UserID: "test-user-list",
		Type:   "list-type",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(5))
}

func TestStore_ListDagInstance(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag := &entity.Dag{
		UserID: "test-user-list-dagins",
		Name:   "test-dag-list-dagins",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	dagID, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		dagIns := &entity.DagInstance{
			DagID:   dagID,
			Trigger: entity.Trigger{Type: entity.DagTriggerManual},
			UserID:  "test-user-list-dagins",
			Status:  entity.DagInstanceStatusInit,
		}
		_, err := testStore.CreateDagIns(ctx, dagIns)
		require.NoError(t, err)
	}

	list, err := testStore.ListDagInstance(ctx, &mod.ListDagInstanceInput{
		DagIDs: []string{dagID},
		Limit:  10,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 3)

	count, err := testStore.GetDagInstanceCount(ctx, map[string]interface{}{
		"dag_id": dagID,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(3))
}

func TestStore_ListTaskInstance(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag := &entity.Dag{
		UserID: "test-user-list-task",
		Name:   "test-dag-list-task",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	dagID, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	dagIns := &entity.DagInstance{
		DagID:   dagID,
		Trigger: entity.Trigger{Type: entity.DagTriggerManual},
		UserID:  "test-user-list-task",
		Status:  entity.DagInstanceStatusInit,
	}
	dagInsID, err := testStore.CreateDagIns(ctx, dagIns)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		taskIns := &entity.TaskInstance{
			DagInsID: dagInsID,
			TaskID:   fmt.Sprintf("task-%d", i),
			Status:   entity.TaskInstanceStatusInit,
		}
		err := testStore.CreateTaskIns(ctx, taskIns)
		require.NoError(t, err)
	}

	list, err := testStore.ListTaskInstance(ctx, &mod.ListTaskInstanceInput{
		DagInsID: dagInsID,
		Limit:    10,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 3)

	count, err := testStore.GetTaskInstanceCount(ctx, map[string]interface{}{
		"dag_ins_id": dagInsID,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(3))
}

func TestStore_Inbox_CRUD(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()
	inbox := &entity.InBox{
		DocID: "doc-1",
		Topic: "test-topic",
	}

	err := testStore.CreateInbox(ctx, inbox)
	assert.NoError(t, err)
	assert.NotEmpty(t, inbox.ID)

	retrieved, err := testStore.GetInbox(ctx, inbox.ID)
	assert.NoError(t, err)
	assert.Equal(t, "doc-1", retrieved.DocID)
	assert.Equal(t, "test-topic", retrieved.Topic)

	list, err := testStore.ListInbox(ctx, &mod.ListInboxInput{
		DocID: "doc-1",
		Limit: 10,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 1)

	err = testStore.DeleteInbox(ctx, []string{inbox.ID})
	assert.NoError(t, err)

	_, err = testStore.GetInbox(ctx, inbox.ID)
	assert.Error(t, err)
}

func TestStore_Outbox_CRUD(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()
	outbox := &entity.OutBox{
		Topic: "test-outbox-topic",
		Message: map[string]interface{}{
			"key": "value",
		},
	}

	err := testStore.CreatOutBoxMessage(ctx, outbox)
	assert.NoError(t, err)
	assert.NotEmpty(t, outbox.ID)

	list, err := testStore.ListOutBoxMessage(ctx, &entity.OutBoxInput{
		ID:    outbox.ID,
		Limit: 10,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 1)

	err = testStore.DeleteOutBoxMessage(ctx, []string{outbox.ID})
	assert.NoError(t, err)

	list, err = testStore.ListOutBoxMessage(ctx, &entity.OutBoxInput{
		ID:    outbox.ID,
		Limit: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(list))
}

func TestStore_Switch(t *testing.T) {
	cleanupTables(t)

	status, err := testStore.GetSwitchStatus()
	assert.NoError(t, err)

	err = testStore.SetSwitchStatus(true)
	assert.NoError(t, err)

	status, err = testStore.GetSwitchStatus()
	assert.NoError(t, err)
	assert.True(t, status)

	err = testStore.SetSwitchStatus(false)
	assert.NoError(t, err)

	status, err = testStore.GetSwitchStatus()
	assert.NoError(t, err)
	assert.False(t, status)
}

func TestStore_DagVersion(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag := &entity.Dag{
		UserID: "test-user-version",
		Name:   "test-dag-version",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	dagID, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	dagVersion := &entity.DagVersion{
		DagID:     dagID,
		VersionID: "v1.0.0",
		Config: map[string]interface{}{
			"name": "test-config",
		},
		SortTime: time.Now().Unix(),
	}

	id, err := testStore.CreateDagVersion(ctx, dagVersion)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	versions, err := testStore.ListDagVersions(ctx, &mod.ListDagVersionInput{
		DagID: dagID,
		Limit: 10,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(versions), 1)

	history, err := testStore.GetHistoryDagByVersionID(ctx, dagID, "v1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, "v1.0.0", history.VersionID)
}

func TestStore_Marshal_Unmarshal(t *testing.T) {
	dag := &entity.Dag{
		UserID: "test-user",
		Name:   "test-dag",
		Tasks:  []entity.Task{{ID: "task-1", Name: "Task 1"}},
	}

	data, err := testStore.Marshal(dag)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var unmarshaled entity.Dag
	err = testStore.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, "test-user", unmarshaled.UserID)
	assert.Equal(t, "test-dag", unmarshaled.Name)
	assert.Equal(t, 1, len(unmarshaled.Tasks))
}

func TestStore_GetDagWithOptionalVersion(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag := &entity.Dag{
		UserID: "test-user-opt-version",
		Name:   "test-dag-opt-version",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	dagID, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	retrieved, err := testStore.GetDagWithOptionalVersion(ctx, dagID, "")
	assert.NoError(t, err)
	assert.Equal(t, "test-dag-opt-version", retrieved.Name)

	dagVersion := &entity.DagVersion{
		DagID:     dagID,
		VersionID: "v2.0.0",
		Config: map[string]interface{}{
			"id":     dagID,
			"userid": "test-user-opt-version",
			"name":   "versioned-dag",
			"tasks":  []interface{}{map[string]interface{}{"id": "task-2"}},
		},
		SortTime: time.Now().Unix(),
	}
	_, err = testStore.CreateDagVersion(ctx, dagVersion)
	require.NoError(t, err)

	retrieved, err = testStore.GetDagWithOptionalVersion(ctx, dagID, "v2.0.0")
	assert.NoError(t, err)
	assert.Equal(t, "versioned-dag", retrieved.Name)
}

func TestStore_DeleteDag(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag := &entity.Dag{
		UserID: "test-user-delete",
		Name:   "test-dag-delete",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	dagID, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	err = testStore.DeleteDag(ctx, dagID)
	assert.NoError(t, err)

	_, err = testStore.GetDag(ctx, dagID)
	assert.Error(t, err)
}

func TestStore_ListExistDagID(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag1 := &entity.Dag{
		UserID: "test-user-exist",
		Name:   "test-dag-exist-1",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	id1, err := testStore.CreateDag(ctx, dag1)
	require.NoError(t, err)

	dag2 := &entity.Dag{
		UserID: "test-user-exist",
		Name:   "test-dag-exist-2",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	id2, err := testStore.CreateDag(ctx, dag2)
	require.NoError(t, err)

	existIDs, err := testStore.ListExistDagID(ctx, []string{id1, id2, "non-existent-id"})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(existIDs))
	assert.Contains(t, existIDs, id1)
	assert.Contains(t, existIDs, id2)
}

func TestStore_UpdateDagIncValue(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag := &entity.Dag{
		UserID: "test-user-inc",
		Name:   "test-dag-inc",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	dagID, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	err = testStore.UpdateDagIncValue(ctx, dagID, "counter", 1)
	assert.NoError(t, err)

	retrieved, err := testStore.GetDag(ctx, dagID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved.IncValues)
	assert.Equal(t, 1, int(retrieved.IncValues["counter"].(float64)))

	err = testStore.UpdateDagIncValue(ctx, dagID, "counter", 2)
	assert.NoError(t, err)

	retrieved, err = testStore.GetDag(ctx, dagID)
	assert.NoError(t, err)
	assert.Equal(t, 2, int(retrieved.IncValues["counter"].(float64)))
}

func TestStore_RetryDagIns(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag := &entity.Dag{
		UserID: "test-user-retry",
		Name:   "test-dag-retry",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	dagID, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	dagIns := &entity.DagInstance{
		DagID:   dagID,
		Trigger: entity.Trigger{Type: entity.DagTriggerManual},
		UserID:  "test-user-retry",
		Status:  entity.DagInstanceStatusFailed,
		Reason:  "original failure",
	}
	dagInsID, err := testStore.CreateDagIns(ctx, dagIns)
	require.NoError(t, err)

	taskIns := &entity.TaskInstance{
		DagInsID: dagInsID,
		TaskID:   "task-1",
		Status:   entity.TaskInstanceStatusFailed,
		Reason:   "task failure",
	}
	err = testStore.CreateTaskIns(ctx, taskIns)
	require.NoError(t, err)

	err = testStore.RetryDagIns(ctx, dagInsID, []string{taskIns.ID})
	assert.NoError(t, err)

	retrievedDagIns, err := testStore.GetDagInstance(ctx, dagInsID)
	assert.NoError(t, err)
	assert.Equal(t, entity.DagInstanceStatusInit, retrievedDagIns.Status)
	assert.Empty(t, retrievedDagIns.Reason)

	retrievedTaskIns, err := testStore.GetTaskIns(ctx, taskIns.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity.TaskInstanceStatusInit, retrievedTaskIns.Status)
	assert.Empty(t, retrievedTaskIns.Reason)
}

func TestStore_GroupDagInstance(t *testing.T) {
	cleanupTables(t)

	ctx := context.Background()

	dag := &entity.Dag{
		UserID: "test-user-group",
		Name:   "test-dag-group",
		Trigger: entity.Trigger{
			Type: entity.DagTriggerManual,
		},
		Status: entity.DagStatusNormal,
		Tasks:  []entity.Task{{ID: "task-1"}},
	}
	dagID, err := testStore.CreateDag(ctx, dag)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		dagIns := &entity.DagInstance{
			DagID:   dagID,
			Trigger: entity.Trigger{Type: entity.DagTriggerManual},
			UserID:  "test-user-group",
			Status:  entity.DagInstanceStatusInit,
		}
		_, err := testStore.CreateDagIns(ctx, dagIns)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		dagIns := &entity.DagInstance{
			DagID:   dagID,
			Trigger: entity.Trigger{Type: entity.DagTriggerManual},
			UserID:  "test-user-group",
			Status:  entity.DagInstanceStatusSuccess,
		}
		_, err := testStore.CreateDagIns(ctx, dagIns)
		require.NoError(t, err)
	}

	groups, err := testStore.GroupDagInstance(ctx, &mod.GroupInput{
		GroupBy: "status",
		Limit:   10,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(groups), 1)
}
