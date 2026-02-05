package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"go.opentelemetry.io/otel/codes"

	"data-model/common"
	"data-model/interfaces"
	dtype "data-model/interfaces/data_type"
	"data-model/logics"
	"data-model/logics/data_view"
	"data-model/logics/permission"
)

var (
	dvmServiceOnce sync.Once
	dvmService     interfaces.DataViewMonitorService
)

type dataViewMonitorService struct {
	appSetting   *common.AppSetting
	dvs          interfaces.DataViewService
	dvgs         interfaces.DataViewGroupService
	ps           interfaces.PermissionService
	dsa          interfaces.DataSourceAccess
	vma          interfaces.VegaMetadataAccess
	lastSyncTime string
	mu           sync.RWMutex
	results      []interfaces.SyncResult
	batchResults []interfaces.BatchResult
	initialized  bool
	batchSize    int
}

func NewDataViewMonitorService(appSetting *common.AppSetting) interfaces.DataViewMonitorService {
	dvmServiceOnce.Do(func() {
		dvmService = &dataViewMonitorService{
			appSetting:   appSetting,
			dvs:          data_view.NewDataViewService(appSetting),
			dvgs:         data_view.NewDataViewGroupService(appSetting),
			ps:           permission.NewPermissionService(appSetting),
			dsa:          logics.DSA,
			vma:          logics.VMA,
			results:      make([]interfaces.SyncResult, 0),
			batchResults: make([]interfaces.BatchResult, 0),
			initialized:  true,
			batchSize:    getDefaultBatchSize(),
			lastSyncTime: "",
		}

		if appSetting.ServerSetting.WatchMetadataEnabled {
			logger.Infof("Watch metadata enabled, Sync service initialized. Will perform full sync on first run")
			// accountID 存入 context 中
			ctx := context.WithValue(context.Background(), interfaces.ACCOUNT_INFO_KEY,
				interfaces.AccountInfo{
					ID:   interfaces.ADMIN_ID,
					Type: interfaces.ACCESSOR_TYPE_USER,
				})

			// 轮询元数据管理接口，常驻进程
			go dvmService.PollingMetadata(ctx)
		} else {
			logger.Infof("Watch metadata disabled, will not start polling metadata table")
		}
	})

	return dvmService
}

// 每隔 1min 轮询元数据管理接口，决定视图同步
func (dvmService *dataViewMonitorService) PollingMetadata(ctx context.Context) {
	interval := dvmService.appSetting.ServerSetting.WatchMetadataInterval
	logger.Infof("Polling metadata table, interval: %v", interval*time.Second)
	for {
		err := dvmService.syncViews(ctx)
		if err != nil {
			logger.Errorf("Error syncing views: %v", err)
		}
		time.Sleep(interval * time.Second)
	}
}

// syncViews 执行视图同步
func (dvmService *dataViewMonitorService) syncViews(ctx context.Context) error {
	if !dvmService.initialized {
		logger.Infof("Service not initialized, skipping sync")
		return nil
	}

	logger.Infof("Starting batch view synchronization...")
	lastSync := dvmService.GetLastSyncTime()

	// 决定同步类型
	if lastSync == "" {
		logger.Infof("Performing full sync - all metadata")
		// 全量同步
		err := dvmService.Sync(ctx, interfaces.SyncType_Full, "")
		if err != nil {
			logger.Errorf("Error performing full sync: %v", err)
			return err
		}
		logger.Infof("Full sync completed successfully")

	} else {
		logger.Infof("Performing incremental sync since: %v", lastSync)
		// 增量同步
		err := dvmService.Sync(ctx, interfaces.SyncType_Incremental, lastSync)
		if err != nil {
			logger.Errorf("Error performing incremental sync: %v", err)
			return err
		}
		logger.Infof("Incremental sync completed successfully")
	}

	return nil
}

// 同步数据，记录每次同步开始的时间为 lastSyncTime
func (dvmService *dataViewMonitorService) Sync(ctx context.Context, syncType string, lastSyncTime string) error {
	// 记录同步的开始时间
	startTime := time.Now()

	// 验证同步类型
	if syncType != interfaces.SyncType_Full && syncType != interfaces.SyncType_Incremental {
		return fmt.Errorf("invalid sync type: %s", syncType)
	}

	logger.Infof("Starting %s sync, lastSyncTime: %s", syncType, lastSyncTime)

	// 按照数据源同步，先获取数据源列表，再循环数据源获取元数据列表
	dataSourceList, err := dvmService.dsa.ListDataSources(ctx)
	if err != nil {
		logger.Errorf("Error getting data source list: %v", err)
		return fmt.Errorf("failed to get data source list: %w", err)
	}
	logger.Infof("Successfully got data source list, count: %d", len(dataSourceList.Entries))

	// 获取内置的分组列表
	builtinGroups, _, err := dvmService.dvgs.ListDataViewGroups(ctx, &interfaces.ListViewGroupQueryParams{
		Builtin: []bool{true},
		PaginationQueryParameters: interfaces.PaginationQueryParameters{
			Offset:    0,
			Limit:     -1,
			Sort:      interfaces.DATA_VIEW_GROUP_SORT[interfaces.DEFAULT_DATA_VIEW_GROUP_SORT],
			Direction: interfaces.ASC_DIRECTION,
		},
	}, false)
	if err != nil {
		logger.Errorf("Error getting builtin data view groups: %v", err)
		return fmt.Errorf("failed to get builtin groups: %w", err)
	}

	err = dvmService.compareDataSourceAndBuiltinGroups(ctx, dataSourceList, builtinGroups)
	if err != nil {
		logger.Errorf("Error comparing data source and builtin groups: %v", err)
		return fmt.Errorf("failed to compare data sources and groups: %w", err)
	}

	// 循环数据源获取元数据列表
	// 记录一个flag，是否全部更新成功
	allSuccess := true
	successCount := 0
	totalCount := len(dataSourceList.Entries) - 1 // 去掉index_base数据源

	for _, dataSource := range dataSourceList.Entries {
		// 跳过index_base数据源
		if dataSource.Type == interfaces.DataSourceType_IndexBase {
			logger.Debugf("Skipping index_base data source: %s", dataSource.Name)
			continue
		}

		// 记录每个数据源的同步时间
		dataSourceStartTime := time.Now()
		err := dvmService.syncDataSource(ctx, dataSource, syncType, lastSyncTime)
		// 记录当前数据源的同步时间
		dataSourceEndTime := time.Now()
		dataSourceDuration := dataSourceEndTime.Sub(dataSourceStartTime).Milliseconds()
		logger.Infof("Data source '%s' synced in %v ms", dataSource.Name, dataSourceDuration)
		if err != nil {
			logger.Errorf("Error syncing data source '%s': %v", dataSource.Name, err)
			allSuccess = false
		} else {
			successCount++
		}
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime).Milliseconds()

	// 记录同步统计信息
	logger.Infof("Sync completed in %v ms, success: %d/%d data sources", duration, successCount, totalCount)

	// 如果全部数据源同步成功，更新 lastSyncTime
	if allSuccess {
		dvmService.setLastSyncTime(startTime.Format(time.RFC3339Nano))
		logger.Infof("Sync state updated in memory: %v", dvmService.GetLastSyncTime())
	} else {
		logger.Warnf("Sync completed with errors, %d/%d data sources failed", totalCount-successCount, totalCount)
	}

	return nil
}

// 同步单个数据源
func (dvmService *dataViewMonitorService) syncDataSource(ctx context.Context, dataSource *interfaces.DataSource, syncType string, lastSyncTime string) error {
	// 先操作标记源表删除，源表删除需要全量获取表，不能拿增量查询的库表和全量视图对比
	// 获取这个数据源（分组）下的所有视图
	dataViews, err := dvmService.dvs.GetDataViewsBySourceID(ctx, dataSource.ID)
	if err != nil {
		return fmt.Errorf("failed to get data views: %w", err)
	}

	// 这个数据源下的所有视图，用技术名称作为key
	dataViewsMap := make(map[string]*interfaces.DataView)
	// 维护业务名称和业务id的map，避免新生成的业务名称会在分组内重复
	// dataViewBusinessNameMap := make(map[string]string)
	for _, dView := range dataViews {
		dataViewsMap[dView.TechnicalName] = dView
		// dataViewBusinessNameMap[dView.ViewName] = dView.ViewID
	}

	// 标记源表被删除的视图
	err = dvmService.markDeletedTablesAsSourceDeleted(ctx, dataSource, dataViewsMap)
	if err != nil {
		return fmt.Errorf("failed to mark deleted tables: %w", err)
	}

	metadataList, err := dvmService.getMetadataBySourceID(ctx, dataSource.ID, lastSyncTime)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	logger.Infof("Found %d metadata records for data source '%s' when %s sync", len(metadataList), dataSource.Name, syncType)

	// 如果是全量同步，且没有元数据，但有视图，需要标记为源表删除
	// 如果是增量同步，没有更新的元数据，不需要向下走 continue
	if len(metadataList) == 0 {
		logger.Infof("No metadata found for data source '%s' when %s sync", dataSource.Name, syncType)

		// 如果是全量同步，检查该数据源下是否有视图，如果有则标记为源表删除
		if syncType == interfaces.SyncType_Full {
			return dvmService.markViewsAsSourceDeletedForEmptyMetadata(ctx, dataSource)
		}
		return nil
	}

	// 处理数据源的元数据
	err = dvmService.processMetadataByDataSource(ctx, metadataList, dataSource, dataViewsMap)
	if err != nil {
		return fmt.Errorf("failed to process metadata: %w", err)
	}

	return nil
}

// 标记源表被删除的视图
func (dvmService *dataViewMonitorService) markDeletedTablesAsSourceDeleted(ctx context.Context, dataSource *interfaces.DataSource, dataViewsMap map[string]*interfaces.DataView) error {
	allMetadata, err := dvmService.getMetadataBySourceID(ctx, dataSource.ID, "")
	if err != nil {
		return fmt.Errorf("failed to get metadata when mark deleted tables: %w", err)
	}

	logger.Infof("data source '%s' metadata records count: %d, data views count: %d",
		dataSource.Name, len(allMetadata), len(dataViewsMap))

	// table.Name 是视图的技术名称
	tablesMap := make(map[string]interfaces.SimpleMetadataTable)
	for _, table := range allMetadata {
		tablesMap[table.Name] = table
	}

	// 源表被删除了，标记为源表删除
	deleteViewIDs := make([]string, 0)
	for techName, vv := range dataViewsMap {
		if _, ok := tablesMap[techName]; !ok {
			deleteViewIDs = append(deleteViewIDs, vv.ViewID)
		}
	}

	if len(deleteViewIDs) == 0 {
		logger.Infof("No views need to be marked as source deleted for data source '%s'", dataSource.Name)
		return nil
	}

	// 更新视图状态为源表删除
	err = dvmService.MarkViewAsSourceDeleted(ctx, deleteViewIDs)
	if err != nil {
		return fmt.Errorf("failed to mark views as source deleted: %w", err)
	}

	logger.Infof("Marked %d views as source deleted for data source '%s'", len(deleteViewIDs), dataSource.Name)
	return nil
}

// 为没有元数据的数据源标记视图为源表删除
func (dvmService *dataViewMonitorService) markViewsAsSourceDeletedForEmptyMetadata(ctx context.Context, dataSource *interfaces.DataSource) error {
	// 获取这个数据源下的所有视图
	dataViews, err := dvmService.dvs.GetDataViewsBySourceID(ctx, dataSource.ID)
	if err != nil {
		return fmt.Errorf("failed to get data views: %w", err)
	}

	if len(dataViews) == 0 {
		logger.Infof("No views found for data source '%s' with empty metadata", dataSource.Name)
		return nil
	}

	// 收集需要标记为源表删除的视图ID
	deleteViewIDs := make([]string, 0, len(dataViews))
	for _, view := range dataViews {
		deleteViewIDs = append(deleteViewIDs, view.ViewID)
	}

	// 标记视图为源表删除
	err = dvmService.MarkViewAsSourceDeleted(ctx, deleteViewIDs)
	if err != nil {
		return fmt.Errorf("failed to mark views as source deleted: %w", err)
	}

	logger.Infof("Marked %d views as source deleted for data source '%s' in full sync", len(deleteViewIDs), dataSource.Name)
	return nil
}

// 对比数据源列表和内置分组列表，删除被删除的分组，更新分组名称
func (dvmService *dataViewMonitorService) compareDataSourceAndBuiltinGroups(ctx context.Context,
	dataSourceList *interfaces.ListDataSourcesResult, builtinGroups []*interfaces.DataViewGroup) error {

	// 维护数据源map
	dataSourceMap := make(map[string]*interfaces.DataSource)
	for _, dataSource := range dataSourceList.Entries {
		// 移除index_base数据源
		if dataSource.Type == interfaces.DataSourceType_IndexBase {
			continue
		}
		dataSourceMap[dataSource.ID] = dataSource
	}

	// 维护分组map
	builtinGroupMap := make(map[string]*interfaces.DataViewGroup)
	for _, group := range builtinGroups {
		// 移除index_base分组
		if group.GroupID == interfaces.GroupID_IndexBase {
			continue
		}
		builtinGroupMap[group.GroupID] = group
	}

	// 对比数据源列表和内置分组列表
	//  - 如果数据源存在，分组不存在，创建这个分组
	//  - 如果数据源不存在，分组存在，标记这个分组和这个分组下的视图为标记删除
	//  - 如果数据源和分组都存在，数据源和分组名称不一致，更新分组名称
	// 先删除分组，再新建，避免旧版本脏数据影响整个流程（比如数据连接删了又新建同名字的）
	for _, group := range builtinGroupMap {
		if _, ok := dataSourceMap[group.GroupID]; !ok {
			// 标记删除分组和分组下的视图
			err := dvmService.dvgs.MarkDataViewGroupDeleted(ctx, group.GroupID, true)
			if err != nil {
				logger.Errorf("Error marking data view group '%s' deleted: %v", group.GroupName, err)
				return err
			}
			logger.Infof("Marked data view group '%s' and its views as deleted as it no longer exists in data source list", group.GroupName)
		}
	}

	for _, dataSource := range dataSourceMap {
		dataSourceName := common.CutStringByCharCount(dataSource.Name, interfaces.MaxLength_ViewGroupName)
		if group, ok := builtinGroupMap[dataSource.ID]; !ok {
			// 创建分组
			_, err := dvmService.dvgs.CreateDataViewGroup(ctx, nil, &interfaces.DataViewGroup{
				GroupID:   dataSource.ID,
				GroupName: dataSourceName,
				Builtin:   true,
			})
			if err != nil {
				logger.Errorf("Error creating data view group %s: %v", dataSource.Name, err)
				return err
			}
			logger.Infof("Created data view group %s for data source '%s'", dataSource.Name, dataSource.Name)
		} else {
			// 更新分组名称
			if group.GroupName != dataSourceName {
				group.GroupName = dataSourceName
				err := dvmService.dvgs.UpdateDataViewGroup(ctx, group)
				if err != nil {
					logger.Errorf("Error updating data view group '%s': %v", group.GroupName, err)
					return err
				}
				logger.Infof("Updated data view group '%s' name to '%s'", group.GroupName, dataSource.Name)
			}
		}
	}

	return nil
}

// 对于每个数据源的元数据，进行批量处理
func (dvmService *dataViewMonitorService) processMetadataByDataSource(ctx context.Context, metadataList []interfaces.SimpleMetadataTable,
	dataSource *interfaces.DataSource, dataViewsMap map[string]*interfaces.DataView) error {
	// 将元数据分批次
	batches := chunkMetadata(metadataList, dvmService.batchSize)
	logger.Infof("Batch size: %d, Split into %d batches", dvmService.batchSize, len(batches))

	var allSyncResults []interfaces.SyncResult
	var processedCount, totalCount int
	// var newSyncTime string

	// 处理每个批次
	for i, batch := range batches {
		batchResult, err := dvmService.processBatch(ctx, i+1, batch,
			dataSource, dataViewsMap)
		if err != nil {
			logger.Errorf("Error processing batch %d: %v", i+1, err)
			return err
		}

		allSyncResults = append(allSyncResults, batchResult.Results...)
		processedCount += batchResult.SuccessCount
		totalCount += batchResult.TotalCount

		// 记录批次结果
		dvmService.mu.Lock()
		dvmService.batchResults = append(dvmService.batchResults, batchResult)
		// 保持批次结果列表大小可控
		if len(dvmService.batchResults) > 100 {
			dvmService.batchResults = dvmService.batchResults[len(dvmService.batchResults)-100:]
		}
		dvmService.mu.Unlock()

		// 对于增量同步，更新最大的更新时间
		// if syncType == interfaces.SyncType_Incremental {
		// 	batchMaxTime := dvmService.findMaxUpdatedTimeInBatch(batch)
		// 	// 只有当newSyncTime不为空时才需要比较
		// 	if newSyncTime != "" && batchMaxTime != "" {
		// 		cpResult, err := common.CompareDateTime(batchMaxTime, newSyncTime)
		// 		if err != nil {
		// 			logger.Errorf("Error comparing datetime in batch %d: %v", i+1, err)
		// 			return err
		// 		}

		// 		if cpResult > 0 {
		// 			newSyncTime = batchMaxTime
		// 		}
		// 	} else if batchMaxTime != "" {
		// 		// 如果newSyncTime为空，直接使用batchMaxTime
		// 		newSyncTime = batchMaxTime
		// 	} else if batchMaxTime == "" {
		// 		// 报错：批次中没有更新时间
		// 		logger.Errorf("Batch %d: No updated time found in batch", i+1)
		// 		return fmt.Errorf("batch %d: no updated time found in batch", i+1)
		// 	}
		// }

		logger.Infof("Batch %d/%d completed: %d success, %d errors, created %d views, updated %d views",
			i+1, len(batches), batchResult.SuccessCount, batchResult.ErrorCount, batchResult.NeedCreatedCount,
			batchResult.NeedUpdatedCount)
	}

	// 更新内存中的同步状态
	dvmService.mu.Lock()
	dvmService.results = append(dvmService.results, allSyncResults...)
	// 保持结果列表大小可控
	if len(dvmService.results) > 1000 {
		dvmService.results = dvmService.results[len(dvmService.results)-1000:]
	}
	dvmService.mu.Unlock()

	logger.Infof("Batch synchronization completed. Total: %d batches, %d/%d successful",
		len(batches), processedCount, totalCount)

	return nil
}

// processBatch 处理单个批次
func (dvmService *dataViewMonitorService) processBatch(ctx context.Context, batchID int, metadataList []interfaces.SimpleMetadataTable,
	dataSource *interfaces.DataSource, dataViewsMap map[string]*interfaces.DataView) (interfaces.BatchResult, error) {
	batchStart := time.Now()

	logger.Infof("Processing batch %d with %d metadata records for data source '%s'",
		batchID, len(metadataList), dataSource.Name)

	ids := make([]string, 0, len(metadataList))
	for _, metaTable := range metadataList {
		ids = append(ids, metaTable.ID)
	}

	// 根据id查询元表信息
	metaTablesInfo, err := dvmService.vma.GetMetadataTablesByIDs(ctx, ids)
	if err != nil {
		logger.Errorf("get meta data tables failed, err: %v", err)
		return interfaces.BatchResult{}, err
	}

	// 过滤掉table为null的记录, 并且分组为待创建的视图和待更新的视图
	needCreatedTables := make([]interfaces.MetadataTable, 0, len(metadataList))
	needUpdatedTables := make([]interfaces.MetadataTable, 0, len(metadataList))
	validMetaTableCount := 0
	for _, table := range metaTablesInfo {
		if table.Table != nil {
			validMetaTableCount++
			if _, ok := dataViewsMap[table.Table.Name]; !ok {
				needCreatedTables = append(needCreatedTables, table)
			} else {
				needUpdatedTables = append(needUpdatedTables, table)
			}
		} else {
			logger.Warnf("meta table id '%s', name '%s' in data source '%s' fields are empty, skip",
				table.TableID, table.Table.Name, dataSource.Name)
		}
	}

	batchResult := interfaces.BatchResult{
		BatchID:               batchID,
		TotalMetaTableCount:   len(metadataList),
		InvalidMetaTableCount: len(metadataList) - validMetaTableCount,
		TotalCount:            validMetaTableCount,
		NeedCreatedCount:      len(needCreatedTables),
		NeedUpdatedCount:      len(needUpdatedTables),
		StartTime:             batchStart,
	}

	// 批量创建视图
	err = dvmService.createViews(ctx, needCreatedTables, dataSource)
	if err != nil {
		logger.Errorf("create views failed, err: %v", err)
		return interfaces.BatchResult{}, err
	}

	// 批量更新视图
	err = dvmService.updateViews(ctx, needUpdatedTables, dataViewsMap)
	if err != nil {
		logger.Errorf("update views failed, err: %v", err)
		return interfaces.BatchResult{}, err
	}

	// 记录成功处理数量
	batchResult.SuccessCount = validMetaTableCount
	batchResult.ErrorCount = 0
	batchResult.EndTime = time.Now()

	return batchResult, nil
}

// GetMetadata 获取元数据表
func (dvmService *dataViewMonitorService) getMetadataBySourceID(ctx context.Context, dataSourceID string,
	lastSyncTime string) ([]interfaces.SimpleMetadataTable, error) {
	params := &interfaces.ListMetadataTablesParams{
		DataSourceId: dataSourceID,
		PaginationQueryParameters: interfaces.PaginationQueryParameters{
			Offset: 0,
			Limit:  -1,
		},
		UpdateTime: lastSyncTime,
	}
	// 获取需要同步的元数据
	metadataList, err := dvmService.vma.ListMetadataTablesBySourceID(ctx, params)
	if err != nil {
		logger.Errorf("list meta data tables failed, err: %v", err)
		return nil, err
	}

	return metadataList, nil
}

// createViews 创建视图
func (dvmService *dataViewMonitorService) createViews(ctx context.Context, tables []interfaces.MetadataTable,
	dataSource *interfaces.DataSource) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create data views")
	defer span.End()

	// 使用数据连接的创建者作为视图的创建者
	accountInfo := interfaces.AccountInfo{
		ID:   dataSource.CreatorID,
		Type: dataSource.CreatorType,
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	createViews, err := dvmService.initCreatedViews(tables)
	if err != nil {
		logger.Errorf("init views failed, err: %v", err)
		return err
	}

	logger.Infof("[SyncAtomicView] create view count: %d", len(createViews))
	if _, err = dvmService.dvs.CreateDataViews(ctx, createViews, interfaces.ImportMode_Normal, false); err != nil {
		errDetails := fmt.Sprintf("[SyncAtomicView] create view database error, %v", err)
		span.SetStatus(codes.Error, "CreateView create view database error")
		o11y.Error(ctx, errDetails)
		logger.Errorf(errDetails)

		return errors.New(errDetails)
	}

	span.SetStatus(codes.Ok, "Batch create data views success")
	return nil
}

// 批量更新视图
func (dvmService *dataViewMonitorService) updateViews(ctx context.Context, tables []interfaces.MetadataTable,
	dataViewsMap map[string]*interfaces.DataView) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Atomic View Sync: Update data views")
	defer span.End()

	updateViews, err := dvmService.initUpdatedViews(tables, dataViewsMap)
	if err != nil {
		logger.Errorf("init views failed, err: %v", err)
		return err
	}

	logger.Infof("[SyncAtomicView] update view count: %d", len(updateViews))
	for _, view := range updateViews {
		if err := dvmService.dvs.UpdateDataViewInternal(ctx, view); err != nil {
			span.SetStatus(codes.Error, "Update data view failed")
			o11y.Error(ctx, fmt.Sprintf("Update data view error, %v", err))
			return fmt.Errorf("[SyncAtomicView] update view error, %v", err)
		}
	}

	span.SetStatus(codes.Ok, "Batch update data views success")
	return nil
}

// initCreatedViews 初始化需要创建的视图
func (dvmService *dataViewMonitorService) initCreatedViews(tables []interfaces.MetadataTable) ([]*interfaces.DataView, error) {
	atomicViews := make([]*interfaces.DataView, 0, len(tables))
	for _, table := range tables {
		var selectFields string
		fields := make([]*interfaces.ViewField, len(table.FieldList))
		primaryKeys := make([]string, 0)
		// fieldDisplayNameMap := make(map[string]struct{})
		for i, mField := range table.FieldList {
			fieldType := mField.AdvancedParams.GetValue(interfaces.FieldAdvancedParams_VirtualDataType).(string)
			isNullable := mField.AdvancedParams.GetValue(interfaces.FieldAdvancedParams_IsNullable).(string)
			if fieldType == "" {
				logger.Warnf("table '%s' field '%s' type is empty", table.Table.Name, mField.FieldName)
			}

			var features []interfaces.FieldFeature
			if isOpenSearchOrIndexBaseDataSource(table.DataSource.Type) {
				features = generateNativeFieldFeatures(fieldType, mField)
			}

			fields[i] = &interfaces.ViewField{
				Name:         mField.FieldName,
				DisplayName:  mField.FieldName,
				OriginalName: mField.FieldName,
				Comment:      common.CutStringByCharCount(mField.FieldComment, interfaces.MaxLength_ViewFieldComment),
				Status:       interfaces.ViewScanStatus_New,
				IsPrimaryKey: sql.NullBool{Bool: mField.AdvancedParams.IsPrimaryKey(), Valid: true},
				Type:         fieldType,
				DataLength:   mField.FieldLength,
				DataAccuracy: mField.FieldPrecision,
				IsNullable:   isNullable,
				Features:     features,
			}

			if fieldType == "" { //不支持的类型设置状态
				fields[i].Status = interfaces.FieldScanStatus_NotSupport
			} else {
				selectFields = common.CE(selectFields == "", common.QuotationMark(mField.FieldName),
					fmt.Sprintf("%s,%s", selectFields, common.QuotationMark(mField.FieldName))).(string)
			}

			if mField.AdvancedParams.IsPrimaryKey() {
				primaryKeys = append(primaryKeys, mField.FieldName)
			}
		}

		// 获取 excel 数据源的excelConfig
		var excelFileName string
		var excelConfig *interfaces.ExcelConfig
		if table.DataSource.Type == interfaces.DataSourceType_Excel {
			tableAdvanced := table.Table.AdvancedParams
			excelConfig = &interfaces.ExcelConfig{
				Sheet:            tableAdvanced.GetValue(interfaces.TableAdvancedParams_ExcelSheet).(string),
				StartCell:        tableAdvanced.GetValue(interfaces.TableAdvancedParams_ExcelStartCell).(string),
				EndCell:          tableAdvanced.GetValue(interfaces.TableAdvancedParams_ExcelEndCell).(string),
				HasHeaders:       tableAdvanced.GetValue(interfaces.TableAdvancedParams_ExcelHasHeaders).(bool),
				SheetAsNewColumn: tableAdvanced.GetValue(interfaces.TableAdvancedParams_ExcelSheetAsNewColumn).(bool),
			}
			excelFileName = tableAdvanced.GetValue(interfaces.TableAdvancedParams_ExcelFileName).(string)
		}

		// 如果数据源类型是opensearch，query_type为DSL
		var queryType string
		if table.DataSource.Type == interfaces.DataSourceType_OpenSearch {
			queryType = interfaces.QueryType_DSL
		} else {
			queryType = interfaces.QueryType_SQL
		}

		view := &interfaces.DataView{
			SimpleDataView: interfaces.SimpleDataView{
				ViewID:         table.TableID, // 使用元数据表id
				TechnicalName:  table.Table.Name,
				ViewName:       table.Table.Name,
				Builtin:        true,
				GroupID:        table.DataSource.DataSourceID,
				GroupName:      table.DataSource.DataSourceName,
				Type:           interfaces.ViewType_Atomic,
				QueryType:      queryType,
				DataSourceID:   table.DataSource.DataSourceID,
				DataSourceType: table.DataSource.Type,
				FileName:       excelFileName,
				Status:         interfaces.ViewScanStatus_New,
				Comment:        common.CutStringByCharCount(table.Table.Description, interfaces.MaxLength_ViewComment),
			},
			ExcelConfig:    excelConfig,
			Fields:         fields,
			MetadataFormID: table.TableID,
			PrimaryKeys:    primaryKeys,
		}

		catalogName := table.DataSource.Catalog
		schemaName := table.DataSource.Schema
		// 先用schema，没有再用database
		if schemaName == "" {
			schemaName = table.DataSource.Database
		}

		// 补齐 sqlstr 和 metatable name
		metaTableName := fmt.Sprintf("%s.%s.%s", catalogName, common.QuotationMark(schemaName), common.QuotationMark(table.Table.Name))
		view.SQLStr = fmt.Sprintf("SELECT * FROM %s", metaTableName)
		view.MetaTableName = metaTableName

		atomicViews = append(atomicViews, view)

	}

	return atomicViews, nil
}

// initUpdatedViews 初始化需要更新的视图, 实现“增量更新”且“不覆盖用户自定义配置”
func (dvmService *dataViewMonitorService) initUpdatedViews(tables []interfaces.MetadataTable,
	dataViewsMap map[string]*interfaces.DataView) ([]*interfaces.DataView, error) {
	atomicViews := make([]*interfaces.DataView, 0, len(tables))
	for _, table := range tables {
		existingView, ok := dataViewsMap[table.Table.Name]
		if !ok {
			continue
		}

		// 已有的字段列表
		existingFieldsMap := make(map[string]*interfaces.ViewField)
		for _, vfield := range existingView.Fields {
			existingFieldsMap[vfield.OriginalName] = vfield
		}
		logger.Debugf("update view metadata table %s fields count is %d, view %s fields count is %d",
			table.Table.Name, len(table.FieldList), existingView.ViewName, len(existingView.Fields))

		var selectFields string
		// fields := make([]*interfaces.ViewField, len(table.FieldList))
		primaryKeys := make([]string, 0)
		final_view_fields := make([]*interfaces.ViewField, 0, len(table.FieldList))
		for _, mField := range table.FieldList {
			if oldField, ok := existingFieldsMap[mField.FieldName]; !ok {
				//field new
				logger.Debugf("update view, table name: %s, field name: %s, field not exist in view",
					table.Table.Name, mField.FieldName)
				fieldType := mField.AdvancedParams.GetValue(interfaces.FieldAdvancedParams_VirtualDataType).(string)
				isNullable := mField.AdvancedParams.GetValue(interfaces.FieldAdvancedParams_IsNullable).(string)
				if fieldType == "" {
					logger.Warnf("table '%s' field '%s' type is empty", table.Table.Name, mField.FieldName)
				}

				var osFeatures []interfaces.FieldFeature
				if isOpenSearchOrIndexBaseDataSource(table.DataSource.Type) {
					osFeatures = generateNativeFieldFeatures(fieldType, mField)
				}

				newField := &interfaces.ViewField{
					Name:         mField.FieldName,
					DisplayName:  mField.FieldName,
					OriginalName: mField.FieldName,
					Comment:      common.CutStringByCharCount(mField.FieldComment, interfaces.MaxLength_ViewFieldComment),
					Status:       interfaces.FieldScanStatus_New,
					IsPrimaryKey: sql.NullBool{Bool: mField.AdvancedParams.IsPrimaryKey(), Valid: true},
					Type:         fieldType,
					DataLength:   mField.FieldLength,
					DataAccuracy: mField.FieldPrecision,
					IsNullable:   isNullable,
					Features:     osFeatures,
				}

				if fieldType == "" { //不支持的类型设置状态
					newField.Status = interfaces.FieldScanStatus_NotSupport
				} else {
					selectFields = common.CE(selectFields == "", common.QuotationMark(mField.FieldName),
						fmt.Sprintf("%s,%s", selectFields, common.QuotationMark(mField.FieldName))).(string)
				}
				// 新增字段加入最终字段列表
				final_view_fields = append(final_view_fields, newField)

				if mField.AdvancedParams.IsPrimaryKey() {
					primaryKeys = append(primaryKeys, mField.FieldName)
				}
			} else {
				// field update
				logger.Debugf("update view, table name: %s, field name: %s, field already exists in view",
					table.Table.Name, mField.FieldName)
				fieldType := mField.AdvancedParams.GetValue(interfaces.FieldAdvancedParams_VirtualDataType).(string)
				isNullable := mField.AdvancedParams.GetValue(interfaces.FieldAdvancedParams_IsNullable).(string)
				if fieldType == "" {
					logger.Warnf("table '%s' field '%s' type is empty", table.Table.Name, mField.FieldName)
				}

				var mergedFeatures []interfaces.FieldFeature
				if isOpenSearchOrIndexBaseDataSource(table.DataSource.Type) {
					osFeatures := generateNativeFieldFeatures(fieldType, mField)
					mergedFeatures = MergeFeaturesOptimized(oldField.Features, osFeatures)
				}

				updateField := &interfaces.ViewField{
					Name:         mField.FieldName,
					DisplayName:  oldField.DisplayName,
					OriginalName: mField.FieldName,
					Comment:      oldField.Comment,
					Status:       interfaces.FieldScanStatus_Modify,
					IsPrimaryKey: sql.NullBool{Bool: mField.AdvancedParams.IsPrimaryKey(), Valid: true},
					Type:         fieldType,
					DataLength:   mField.FieldLength,
					DataAccuracy: mField.FieldPrecision,
					IsNullable:   isNullable,
					Features:     mergedFeatures,
				}

				if fieldType == "" { //不支持的类型设置状态
					updateField.Status = interfaces.FieldScanStatus_NotSupport
				} else {
					selectFields = common.CE(selectFields == "", common.QuotationMark(mField.FieldName),
						fmt.Sprintf("%s,%s", selectFields, common.QuotationMark(mField.FieldName))).(string)
				}
				// 更新的字段加入最终字段列表
				final_view_fields = append(final_view_fields, updateField)

				if mField.AdvancedParams.IsPrimaryKey() {
					primaryKeys = append(primaryKeys, mField.FieldName)
				}
			}
		}

		// 获取 excel 数据源的excelConfig
		var excelFileName string
		var excelConfig *interfaces.ExcelConfig
		if table.DataSource.Type == interfaces.DataSourceType_Excel {
			tableAdvanced := table.Table.AdvancedParams
			excelConfig = &interfaces.ExcelConfig{
				Sheet:            tableAdvanced.GetValue(interfaces.TableAdvancedParams_ExcelSheet).(string),
				StartCell:        tableAdvanced.GetValue(interfaces.TableAdvancedParams_ExcelStartCell).(string),
				EndCell:          tableAdvanced.GetValue(interfaces.TableAdvancedParams_ExcelEndCell).(string),
				HasHeaders:       tableAdvanced.GetValue(interfaces.TableAdvancedParams_ExcelHasHeaders).(bool),
				SheetAsNewColumn: tableAdvanced.GetValue(interfaces.TableAdvancedParams_ExcelSheetAsNewColumn).(bool),
			}
			excelFileName = tableAdvanced.GetValue(interfaces.TableAdvancedParams_ExcelFileName).(string)
		}

		// 如果数据源类型是opensearch，query_type为DSL
		var queryType string
		if table.DataSource.Type == interfaces.DataSourceType_OpenSearch {
			queryType = interfaces.QueryType_DSL
		} else {
			queryType = interfaces.QueryType_SQL
		}

		view := &interfaces.DataView{
			SimpleDataView: interfaces.SimpleDataView{
				ViewID:         table.TableID, // 使用元数据表id
				TechnicalName:  table.Table.Name,
				ViewName:       existingView.ViewName,
				Builtin:        true,
				GroupID:        table.DataSource.DataSourceID,
				GroupName:      table.DataSource.DataSourceName,
				Type:           interfaces.ViewType_Atomic,
				QueryType:      queryType,
				DataSourceID:   table.DataSource.DataSourceID,
				DataSourceType: table.DataSource.Type,
				FileName:       excelFileName,
				Status:         interfaces.ViewScanStatus_Modify,
				// Comment:        common.CutStringByCharCount(table.Table.Description, interfaces.CommentCharCountLimit),
				Comment: existingView.Comment,
			},
			ExcelConfig: excelConfig,
			// Fields:         fields,
			Fields:         final_view_fields,
			MetadataFormID: table.TableID,
			PrimaryKeys:    primaryKeys,
		}

		catalogName := table.DataSource.Catalog
		schemaName := table.DataSource.Schema
		// 先用schema，没有再用database
		if schemaName == "" {
			schemaName = table.DataSource.Database
		}

		// 补齐 sqlstr 和 metatable name
		metaTableName := fmt.Sprintf("%s.%s.%s", catalogName, common.QuotationMark(schemaName), common.QuotationMark(table.Table.Name))
		view.SQLStr = fmt.Sprintf("SELECT * FROM %s", metaTableName)
		view.MetaTableName = metaTableName

		atomicViews = append(atomicViews, view)

	}

	return atomicViews, nil
}

// chunkMetadata 将元数据列表分批次
func chunkMetadata(metadataList []interfaces.SimpleMetadataTable, batchSize int) [][]interfaces.SimpleMetadataTable {
	if batchSize <= 0 {
		batchSize = 1000 // 默认批次大小
	}

	var chunks [][]interfaces.SimpleMetadataTable
	for i := 0; i < len(metadataList); i += batchSize {
		end := i + batchSize
		if end > len(metadataList) {
			end = len(metadataList)
		}
		chunks = append(chunks, metadataList[i:end])
	}
	return chunks
}

// getDefaultBatchSize 获取默认批次大小
func getDefaultBatchSize() int {
	return 1000
}

func (dvmService *dataViewMonitorService) MarkViewAsSourceDeleted(ctx context.Context, viewsIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Mark view as source deleted")
	defer span.End()

	logger.Infof("MarkViewAsSourceDeleted views %+v", viewsIDs)
	if len(viewsIDs) == 0 {
		span.SetStatus(codes.Ok, "viewsIDs is empty")
		return nil
	}

	//  更新视图状态为源表删除
	if err := dvmService.dvs.MarkDataViewsDeleted(ctx, nil, viewsIDs); err != nil {
		span.SetStatus(codes.Error, "mark view as source deleted failed")
		o11y.Error(ctx, "mark view as source deleted failed")
		return err
	}

	span.SetStatus(codes.Ok, "mark view as source deleted success")
	return nil
}

// GetLastSyncTime 获取最后同步时间
func (dvmService *dataViewMonitorService) GetLastSyncTime() string {
	dvmService.mu.RLock()
	defer dvmService.mu.RUnlock()
	return dvmService.lastSyncTime
}

// setLastSyncTime 设置最后同步时间
func (dvmService *dataViewMonitorService) setLastSyncTime(lastSyncTime string) {
	dvmService.mu.Lock()
	defer dvmService.mu.Unlock()
	dvmService.lastSyncTime = lastSyncTime
}

// isOpenSearchOrIndexBaseDataSource 判断是否是opensearch或index_base数据源
func isOpenSearchOrIndexBaseDataSource(dataSourceType string) bool {
	return dataSourceType == interfaces.DataSourceType_OpenSearch || dataSourceType == interfaces.DataSourceType_IndexBase
}

// 扫描同步初始化字段特征
func generateNativeFieldFeatures(fieldType string, metaField *interfaces.MetaField) []interfaces.FieldFeature {
	advancedParams := metaField.AdvancedParams
	features := []interfaces.FieldFeature{}
	mappingConfig := advancedParams.GetValue(interfaces.FieldAdvancedParams_MappingConfig).(map[string]any)

	// 类型本身具有的特征
	// 1. 全文特征
	if fieldType == dtype.DataType_Text {
		analyzer, _ := common.GetWithDefault(mappingConfig, interfaces.FieldProperty_Analyzer, "standard")
		features = append(features, interfaces.FieldFeature{
			FeatureName: fmt.Sprintf("autoFulltext_%s", metaField.FieldName),
			FeatureType: interfaces.FieldFeatureType_Fulltext,
			Comment:     "自动同步生成的全文检索特征",
			RefField:    metaField.FieldName,
			IsDefault:   false,
			IsNative:    true,
			Config:      map[string]any{interfaces.FieldProperty_Analyzer: analyzer},
		})
	}

	// 2. 精确匹配特征
	if fieldType == dtype.DataType_String {
		fieldsKeywordIgnoreAbove, _ := common.GetWithDefault(mappingConfig, interfaces.FieldProperty_IgnoreAbove, 256)
		features = append(features, interfaces.FieldFeature{
			FeatureName: fmt.Sprintf("autoKeyword_%s", metaField.FieldName),
			FeatureType: interfaces.FieldFeatureType_Keyword,
			Comment:     "自动同步生成的精确匹配特征",
			RefField:    metaField.FieldName,
			IsDefault:   false,
			IsNative:    true,
			Config:      map[string]any{interfaces.FieldProperty_IgnoreAbove: fieldsKeywordIgnoreAbove},
		})
	}

	// 3. 向量特征
	if fieldType == dtype.DataType_Vector {
		dimension, _ := common.GetWithDefault(mappingConfig, interfaces.FieldProperty_Dimension, 768)
		features = append(features, interfaces.FieldFeature{
			FeatureName: fmt.Sprintf("autoVector_%s", metaField.FieldName),
			FeatureType: interfaces.FieldFeatureType_Vector,
			Comment:     "自动同步生成的向量特征",
			RefField:    metaField.FieldName,
			IsDefault:   false,
			IsNative:    true,
			Config:      map[string]any{interfaces.FieldProperty_Dimension: dimension},
		})
	}

	// 子字段
	if subFields, ok := mappingConfig[interfaces.FieldProperty_Fields].(map[string]any); ok {
		for subFieldName, subField := range subFields {
			var fture interfaces.FieldFeature
			if subFieldMap, ok := subField.(map[string]any); ok {
				subFieldType := subFieldMap[interfaces.FieldProperty_Type].(string)
				switch subFieldType {
				case dtype.IndexBase_DataType_Keyword:
					fieldsKeywordIgnoreAbove, _ := common.GetWithDefault(subFieldMap,
						interfaces.FieldProperty_IgnoreAbove, 256)
					fture = interfaces.FieldFeature{
						FeatureName: fmt.Sprintf("autoKeyword_%s.%s", metaField.FieldName, subFieldName),
						FeatureType: interfaces.FieldFeatureType_Keyword,
						Comment:     "自动同步生成的子字段精确匹配特征",
						RefField:    fmt.Sprintf("%s.%s", metaField.FieldName, subFieldName),
						IsDefault:   false,
						IsNative:    true,
						Config:      map[string]any{interfaces.FieldProperty_IgnoreAbove: fieldsKeywordIgnoreAbove},
					}
				case dtype.IndexBase_DataType_Text:
					analyzer, _ := common.GetWithDefault(subFieldMap, interfaces.FieldProperty_Analyzer, "standard")
					fture = interfaces.FieldFeature{
						FeatureName: fmt.Sprintf("autoFulltext_%s.%s", metaField.FieldName, subFieldName),
						FeatureType: interfaces.FieldFeatureType_Fulltext,
						Comment:     "自动同步生成的子字段全文检索特征",
						RefField:    fmt.Sprintf("%s.%s", metaField.FieldName, subFieldName),
						IsDefault:   false,
						IsNative:    true,
						Config:      map[string]any{interfaces.FieldProperty_Analyzer: analyzer},
					}
				case dtype.IndexBase_DataType_KNNVector:
					dimension, _ := common.GetWithDefault(subFieldMap, interfaces.FieldProperty_Dimension, 768)
					fture = interfaces.FieldFeature{
						FeatureName: fmt.Sprintf("autoVector_%s.%s", metaField.FieldName, subFieldName),
						FeatureType: interfaces.FieldFeatureType_Vector,
						Comment:     "自动同步生成的子字段向量特征",
						RefField:    fmt.Sprintf("%s.%s", metaField.FieldName, subFieldName),
						IsDefault:   false,
						IsNative:    true,
						Config:      map[string]any{interfaces.FieldProperty_Dimension: dimension},
					}
				}
			}
			features = append(features, fture)
		}
	}

	// 启用每个类型的第一个特征
	enableFirstFeatureOfEachType(features)

	return features
}

// 启用每个类型的第一个特征
func enableFirstFeatureOfEachType(features []interfaces.FieldFeature) {
	// 使用map记录每种类型是否已经设置了默认特征
	typeProcessed := make(map[interfaces.FieldFeatureType]bool)

	for i := range features {
		featureType := features[i].FeatureType

		if !typeProcessed[featureType] {
			features[i].IsDefault = true
			typeProcessed[featureType] = true
		}
	}
}

// MergeFeaturesOptimized 执行优化后的合并逻辑：Add + Delete + Patch(Config)
func MergeFeaturesOptimized(dbFeatures []interfaces.FieldFeature, osFeatures []interfaces.FieldFeature) []interfaces.FieldFeature {
	result := make([]interfaces.FieldFeature, 0)

	// dbNativeMap 用于存放数据库中已有的原生特征
	// Key 生成逻辑：Type + RefField (物理指纹)
	dbNativeMap := make(map[string]interfaces.FieldFeature)
	usedNames := make(map[string]bool) // 记录已占用的名称
	var customFeatures []interfaces.FieldFeature

	// 1. 预处理：先锁定所有 CUSTOM 特征和已存在的 NATIVE 特征名称
	for _, f := range dbFeatures {
		usedNames[f.FeatureName] = true
		if f.IsNative {
			// 指纹：Type + RefField
			fingerprint := fmt.Sprintf("%s|%s", f.FeatureType, f.RefField)
			dbNativeMap[fingerprint] = f
		} else {
			customFeatures = append(customFeatures, f)
		}
	}

	// 标记哪些指纹在同步中依然有效
	activeFingerprints := make(map[string]bool)

	// 1. 处理 OpenSearch 最新生成的原生特征
	for _, osF := range osFeatures {
		fingerprint := fmt.Sprintf("%s|%s", osF.FeatureType, osF.RefField)

		if existing, exists := dbNativeMap[fingerprint]; exists {
			// 保留用户修改后的 Name, Comment和启用状态,更新物理层面的 Config (分词器、ignore_above 等)
			if !reflect.DeepEqual(existing.Config, osF.Config) {
				logger.Infof("检测到特征 [%s] 的物理配置变更，已同步新参数", existing.FeatureName)
				existing.Config = osF.Config
			}

			result = append(result, existing)
		} else {
			// 指纹不存在，说明是新增物理映射，执行 ADD, 确保特征名称不重复
			baseName := osF.FeatureName
			osF.FeatureName = GenerateUniqueName(baseName, usedNames)
			result = append(result, osF)
		}
		activeFingerprints[fingerprint] = true
	}

	// 2. 补回自定义特征
	result = append(result, customFeatures...)

	// 3. 消失的指纹不进入 result，即实现 DELETE
	// 数据库中原有的原生特征，如果不在 activeFingerprints 里，就不再加入 result

	return result
}

// GenerateUniqueName 确保名称在字段内不重复
func GenerateUniqueName(baseName string, usedNames map[string]bool) string {
	if !usedNames[baseName] {
		usedNames[baseName] = true
		return baseName
	}

	// 如果重复，尝试加后缀 2, 3...
	for i := 2; ; i++ {
		newName := fmt.Sprintf("%s_%d", baseName, i)
		if !usedNames[newName] {
			usedNames[newName] = true
			return newName
		}
	}
}

// // findMaxUpdatedTimeInBatch 在批次中找到最大的更新时间
// func (dvmService *dataViewMonitorService) findMaxUpdatedTimeInBatch(metadataList []interfaces.SimpleMetadataTable) string {
// 	if len(metadataList) == 0 {
// 		return ""
// 	}

// 	maxTime := metadataList[0].UpdateTime
// 	for _, meta := range metadataList {
// 		cmpResult, err := common.CompareDateTime(meta.UpdateTime, maxTime)
// 		if err != nil {
// 			logger.Errorf("CompareDateTime error: %v", err)
// 			continue
// 		}
// 		if cmpResult > 0 {
// 			maxTime = meta.UpdateTime
// 		}
// 	}
// 	return maxTime
// }

// // dataViewBusinessNameMap 用于记录已存在的业务名称，避免重复, 以数据源为单位
// func (dvmService *dataViewMonitorService) AutomaticallyForm(ctx context.Context, table *interfaces.TableInfo,
// 	dataViewBusinessNameMap map[string]string) string {
// 	/*
// 		表业务名称按以下顺序自动生成：
// 		    来自加工模型关联的业务表名称
// 		    表注释
// 		    数据理解
// 		    表技术名称
// 	*/

// 	cleanedDescription := CleanDisplayName(table.Description)

// 	businessName := common.CutStringByCharCount(cleanedDescription, interfaces.BusinessNameCharCountLimit)
// 	if businessName == "" {
// 		businessName = common.CutStringByCharCount(table.Name, interfaces.BusinessNameCharCountLimit)
// 	}

// 	if oldID, ok := dataViewBusinessNameMap[businessName]; ok {
// 		// 如果id变化了且存在重复的视图名称，前面拼接上表技术名称
// 		if oldID != table.ID {
// 			businessName = common.CutStringByCharCount(fmt.Sprintf("%s_%s", table.Name, businessName),
// 				interfaces.BusinessNameCharCountLimit)
// 		}
// 	}

// 	dataViewBusinessNameMap[businessName] = table.ID

// 	return businessName
// }

// func (dvmService *dataViewMonitorService) AutomaticallyField(ctx context.Context, field *interfaces.MetaField,
// 	fieldDisplayNameMap map[string]struct{}) string {
// 	/*
// 		列业务名称按以下顺序自动生成：
// 		    来自加工模型关联的业务表“字段中文名称
// 		    字段注释
// 		    数据理解
// 		    列技术名称
// 	*/

// 	cleanedFieldComment := CleanDisplayName(field.FieldComment)

// 	displayName := common.CutStringByCharCount(cleanedFieldComment, interfaces.BusinessNameCharCountLimit)

// 	if displayName == "" {
// 		displayName = common.CutStringByCharCount(field.FieldName, interfaces.BusinessNameCharCountLimit)
// 	}

// 	if _, ok := fieldDisplayNameMap[displayName]; ok {
// 		// 如果存在重复的字段名称，业务名称前面拼接上字段的原始名称
// 		displayName = common.CutStringByCharCount(fmt.Sprintf("%s_%s", field.FieldName, displayName),
// 			interfaces.BusinessNameCharCountLimit)
// 	}

// 	fieldDisplayNameMap[displayName] = struct{}{}

// 	return displayName
// }

// // CleanDisplayName 清洗显示名称，移除所有不安全字符
// // 只保留：字母、数字、下划线、短横线、点、中文（包含日文、韩文等）
// // 空格会被移除
// func CleanDisplayName(input string) string {
// 	if input == "" {
// 		return input
// 	}

// 	// 构建白名单正则表达式
// 	// \p{L} : 所有字母（包括中文、日文、韩文等）
// 	// \p{N} : 所有数字
// 	// 加上：下划线 _、短横线 -、点 .
// 	pattern := `[^\p{L}\p{N}_\-\.]`

// 	// 编译正则表达式
// 	re := regexp.MustCompile(pattern)

// 	// 移除所有不在白名单中的字符
// 	cleaned := re.ReplaceAllString(input, "")

// 	return cleaned
// }
