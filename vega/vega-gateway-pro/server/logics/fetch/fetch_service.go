package fetch

import (
	"context"
	cryptoRand "crypto/rand"
	"fmt"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/panjf2000/ants/v2"
	mathRand "math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
	"vega-gateway-pro/common"
	"vega-gateway-pro/common/rsa"
	"vega-gateway-pro/interfaces"
	"vega-gateway-pro/logics"
	"vega-gateway-pro/logics/fetch/connectors"
	"vega-gateway-pro/logics/fetch/sqlglot"
	"vega-gateway-pro/version"
)

var (
	fetchServiceOnce sync.Once
	fetchService     interfaces.FetchService
)

type Service struct {
	appSetting           *common.AppSetting
	dataConnectionAccess interfaces.DataConnectionAccess
	vegaCalculateAccess  interfaces.VegaCalculateAccess
	queryPool            *ants.Pool
	queryIdCounter       int
	lastTimestamp        string
	lastTimeInSec        int64
	lastTimeInDay        int64
	coordinatorId        string
	queryCache           sync.Map   // 用于存储查询状态
	queryLocks           sync.Map   // 用于管理每个查询的锁
	QuerySize            int        // 当前查询的数量
	querySizeLock        sync.Mutex // 用于保护QuerySize的锁
}

func NewFetchService(appSetting *common.AppSetting) interfaces.FetchService {
	fetchServiceOnce.Do(func() {
		service := &Service{
			appSetting:           appSetting,
			dataConnectionAccess: logics.DataConnectionAccess,
			vegaCalculateAccess:  logics.VegaCalculateAccess,
			coordinatorId:        generateRandomCoordId(),
		}
		fetchService = service

		service.InitQueryPool(appSetting.PoolSetting)                       // 初始化查询协程池
		go service.startCacheCleaner(appSetting.QuerySetting.CleanInterval) // 启动缓存清理器
	})
	return fetchService
}

// InitQueryPool 初始化查询协程池
func (fs *Service) InitQueryPool(poolSetting common.PoolSetting) {
	pool, err := ants.NewPool(poolSetting.QueryPoolSize, ants.WithPreAlloc(true), ants.WithNonblocking(false))
	if err != nil {
		logger.Errorf("Init query pool failed, %s", err.Error())
		panic(err)
	}

	fs.queryPool = pool
}

// 启动缓存清理器
func (fs *Service) startCacheCleaner(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		logger.Debugf("Clean query cache, query size: %d", fs.QuerySize)
		fs.queryCache.Range(func(key, value interface{}) bool {
			if resultCache, ok := value.(*interfaces.ResultCache); ok && time.Now().After(resultCache.MaxRunTime) {
				fs.cleanQuery(key, resultCache) // 清理查询
			}
			return true
		})
		logger.Debugf("Clean query cache done, query size: %d", fs.QuerySize)
	}
}

// cleanQuery 清理查询
func (fs *Service) cleanQuery(key any, resultCache *interfaces.ResultCache) {

	// 关闭查询结果通道
	if resultCache != nil && resultCache.ResultChan != nil {
		close(resultCache.ResultChan)
	}

	fs.queryCache.Delete(key) // 删除缓存
	fs.queryLocks.Delete(key) // 删除查询锁

	// 减少查询数量
	fs.querySizeLock.Lock()
	fs.QuerySize--
	fs.querySizeLock.Unlock()
}

func (fs *Service) FetchQuery(ctx context.Context, query *interfaces.FetchQueryReq) (resp *interfaces.FetchResp, err error) {

	// 检查查询池是否已满
	fs.querySizeLock.Lock()
	if fs.QuerySize == fs.appSetting.PoolSetting.QueryPoolSize {
		logger.Errorf("Query pool is full, %d", fs.QuerySize)
		fs.querySizeLock.Unlock()
		return nil, rest.NewHTTPError(ctx, http.StatusServiceUnavailable, rest.PublicError_ServiceUnavailable).
			WithErrorDetails("Query pool is full")
	}
	fs.QuerySize++
	fs.querySizeLock.Unlock()

	queryId := fs.generateQueryId()
	slug := generateSlug()

	// 初始化结果通道缓存
	queryCacheKey := fmt.Sprintf("%s_%s", queryId, slug)
	resultChan := make(chan *[]any, fs.appSetting.QuerySetting.DataCacheSize)
	resultCache := &interfaces.ResultCache{
		ResultSet:  nil,
		Token:      0,
		Columns:    nil,
		ResultChan: resultChan,
		Error:      nil,
		MaxRunTime: time.Now().Add(fs.appSetting.QuerySetting.MaxRunTime),
	}
	fs.queryCache.Store(queryCacheKey, resultCache)

	// 记录开始时间
	startTime := time.Now()
	defer func() {

		if r := recover(); r != nil { // 处理查询中的panic
			logger.Errorf("Sql: %s, queryId: %s, slug: %s, query panic in goroutine: %v", query.Sql, queryId, slug, r)
			fs.cleanQuery(queryCacheKey, resultCache)
		} else if ctx.Err() != nil { // 处理查询上下文取消
			logger.Errorf("Sql: %s, queryId: %s, slug: %s, context err: %v", query.Sql, queryId, slug, ctx.Err())
			fs.cleanQuery(queryCacheKey, resultCache)
		} else if err != nil { // 处理查询错误
			logger.Errorf("Sql: %s, queryId: %s, slug: %s, fetch query failed with error: %v", query.Sql, queryId, slug, err)
			fs.cleanQuery(queryCacheKey, resultCache)
		} else if resultCache.ResultSet == nil && len(resultCache.ResultChan) == 0 { // 处理查询完成
			logger.Debugf("Sql: %s, queryId: %s, slug: %s, query completed", query.Sql, queryId, slug)
			fs.cleanQuery(queryCacheKey, resultCache)
		}

		// 记录执行耗时
		logger.Debugf("Fetch query in %v, sql: %s, queryId: %s, slug: %s", time.Since(startTime), query.Sql, queryId, slug)
	}()

	catalog := ""

	logger.Infof("Original SQL: %s", query.Sql)

	// 处理同步查询的LIMIT子句
	if query.Type == 1 &&
		(strings.HasPrefix(strings.ToUpper(query.Sql), "SELECT") ||
			strings.HasPrefix(strings.ToUpper(query.Sql), "WITH")) {
		query.Sql = fmt.Sprintf("SELECT * FROM (%s) AS subquery LIMIT %d", query.Sql, *query.BatchSize)
		logger.Infof("Processed SQL: %s", query.Sql)
	}
	sql := query.Sql

	//提取表名
	tablesResult, err := sqlglot.ExtractTables(sql, "trino")
	if err != nil {
		logger.Errorf("Extract tables failed: %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}
	if len(tablesResult.Tables) == 0 {
		logger.Errorf("Extract tables failed: sql not contain table")
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("Sql not contain table")
	}
	for _, table := range tablesResult.Tables {
		if catalog == "" && table.Catalog != "" {
			catalog = table.Catalog
		}
		if catalog != "" && table.Catalog != "" && table.Catalog != catalog {
			catalog = ""
			break
		}
	}

	if catalog != "" { //单源查询
		logger.Infof("Single-source query, SQL: %s", query.Sql)
		//获取数据源id对应的数据源信息
		dataSource, err := fs.dataConnectionAccess.GetDataSourceById(ctx, query.DataSourceId)
		if err != nil {
			logger.Errorf("Get data source failed: %s", err.Error())
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(err.Error())
		}
		if dataSource == nil {
			logger.Errorf("Get data source failed: data source not found")
			return nil, rest.NewHTTPError(ctx, http.StatusNotFound, rest.PublicError_NotFound).
				WithErrorDetails("data source not found")
		}

		//校验data_source_id和catalog是否匹配
		if dataSource.BinData.CatalogName != catalog {
			logger.Errorf("Catalog name not match: %s != %s", dataSource.BinData.CatalogName, catalog)
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails("catalog name not match")
		}

		dataSource.BinData.Password, err = rsa.Decrypt(dataSource.BinData.Password)
		if err != nil {
			logger.Errorf("Decrypt password failed: %s", err.Error())
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(err.Error())
		}

		//根据数据源类型创建连接器
		connector, err := connectors.NewConnectorHandler(dataSource)
		if err != nil {
			logger.Errorf("New connector handler failed: %s", err.Error())

			if !strings.Contains(err.Error(), "unsupported data source type") {
				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
					WithErrorDetails(err.Error())
			}

			// 数据源类型连接器未适配，查询退化，转发给Etrino
			logger.Infof("Degradation to Etrino, SQL: %s", query.Sql)
			err := fs.getDataFromEtrino(ctx, query.Sql, query.Type, queryId, slug)
			if err != nil {
				return nil, err
			}
		} else {
			//去除sql中的catalog.
			sql = strings.ReplaceAll(sql, fmt.Sprintf("%s.", catalog), "")

			// 直连查询获取结果集
			logger.Infof("Direct connection query, SQL: %s", sql)
			resultSet, err := connector.GetResultSet(sql)
			if err != nil {
				logger.Errorf("Direct connection SQL query failed, %s", err.Error())

				// 方言转化
				sqlParseResult, err := sqlglot.TranspileSQL(sql, "trino", dataSource.Type)
				if err != nil {
					logger.Errorf("Transpile SQL failed: %s", err.Error())
					// 查询退化，转发给Etrino
					logger.Infof("Degradation to Etrino query, SQL: %s", query.Sql)
					err := fs.getDataFromEtrino(ctx, query.Sql, query.Type, queryId, slug)
					if err != nil {
						return nil, err
					}
				} else {
					// 直连查询获取结果集
					logger.Infof("Transpiled SQL query, SQL: %s", sqlParseResult.SQL)
					connector, err = connectors.NewConnectorHandler(dataSource)
					if err != nil {
						logger.Errorf("New connector handler failed: %s", err.Error())

						// 数据源类型连接器未适配，查询退化，转发给Etrino
						logger.Infof("Degradation to Etrino, SQL: %s", query.Sql)
						err := fs.getDataFromEtrino(ctx, query.Sql, query.Type, queryId, slug)
						if err != nil {
							return nil, err
						}
					} else {
						resultSet, err = connector.GetResultSet(sqlParseResult.SQL)
						if err != nil {
							logger.Errorf("Transpiled SQL query failed, %s", err.Error())

							// 查询退化，转发给Etrino
							logger.Infof("Degradation to Etrino query, SQL: %s", query.Sql)
							err := fs.getDataFromEtrino(ctx, query.Sql, query.Type, queryId, slug)
							if err != nil {
								return nil, err
							}
						} else {
							// 直连数据查询
							err = fs.getDataFromQueryResult(ctx, connector, resultSet, query.Type, queryId, slug)
							if err != nil {
								return nil, err
							}
						}
					}
				}
			} else {
				// 直连数据查询
				err = fs.getDataFromQueryResult(ctx, connector, resultSet, query.Type, queryId, slug)
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		//多源查询，转发给Etrino
		logger.Infof("Multi-source query, degradation to Etrino query, SQL: %s", query.Sql)
		err := fs.getDataFromEtrino(ctx, query.Sql, query.Type, queryId, slug)
		if err != nil {
			return nil, err
		}
	}

	// 处理查询结果
	return fs.handleQueryResult(ctx, query.Type, *query.Timeout, *query.BatchSize, queryId, slug, 0)
}

// getDataFromQueryResult 从查询结果集中获取数据
func (fs *Service) getDataFromQueryResult(ctx context.Context, connector connectors.ConnectorHandler, resultSet any, queryType int, queryId string, slug string) error {

	queryCacheKey := fmt.Sprintf("%s_%s", queryId, slug)

	existingCache, ok := fs.queryCache.Load(queryCacheKey)
	if !ok {
		logger.Errorf("Query does not exist, queryId: %s, slug: %s", queryId, slug)
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails("Query does not exist")
	}
	resultCache, _ := existingCache.(*interfaces.ResultCache)

	resultCache.ResultSet = resultSet

	err := fs.queryPool.Submit(func() {

		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("QueryId: %s, slug: %s, direct query panic in goroutine: %v", queryId, slug, r)
				logger.Error(err)
				fs.storeToCache(resultCache, nil, nil, nil, err)
				connector.Close()
			}
			logger.Debugf("QueryId: %s, slug: %s, direct query goroutine exit", queryId, slug)
		}()

		// 获取列信息
		columns, err := connector.GetColumns(resultSet)
		if err != nil {
			fs.storeToCache(resultCache, nil, nil, nil, err)
		}

		for {
			// 如果缓存不存在, 则退出循环
			if _, ok := fs.queryCache.Load(queryCacheKey); !ok {
				connector.Close()
				break
			}

			// 执行查询
			resultSet, data, err := connector.GetData(resultSet, len(columns), queryType, fs.appSetting.QuerySetting.DataQuerySize)

			// 缓存结果
			fs.storeToCache(resultCache, resultSet, columns, data, err)

			// 如果发生错误，退出循环
			if err != nil {
				break
			}

			// query执行完毕，退出循环
			if resultSet == nil {
				break
			}

			// 定期检查缓存中是否有数据
			fs.checkCachePeriodically(resultCache)
		}
	})

	if err != nil {
		logger.Errorf("Get data from query result failed, %s", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}

	return nil
}

// storeToCache 存储查询到缓存中
func (fs *Service) storeToCache(oldResultCache *interfaces.ResultCache, resultSet any, columns []*interfaces.Column, resData []*[]any, err error) {
	if err != nil {
		oldResultCache.Error = err
	} else {
		oldResultCache.Columns = columns
		for _, data := range resData {
			oldResultCache.ResultChan <- data
		}
	}
	oldResultCache.ResultSet = resultSet
}

// checkCachePeriodically 定期检查缓存通道是否有数据
func (fs *Service) checkCachePeriodically(cache *interfaces.ResultCache) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		// 定期检查是否有数据
		if len(cache.ResultChan) < fs.appSetting.QuerySetting.DataQuerySize {
			return
		}
	}
}

// getDataFromEtrino 从Etrino获取数据
func (fs *Service) getDataFromEtrino(ctx context.Context, sql string, queryType int, queryId string, slug string) error {

	statementData, err := fs.vegaCalculateAccess.StatementQuery(ctx, sql)
	if err != nil {
		logger.Errorf("Etrino statement query failed, %s", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}

	if statementData == nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails("StatementData is nil")
	}

	nextUri := statementData.NextUri

	// 第一阶段：处理queued阶段
	for nextUri != "" && strings.Contains(nextUri, "queued") {
		nextData, err := fs.vegaCalculateAccess.NextUriQuery(ctx, nextUri)
		if err != nil {
			logger.Errorf("Etrino queued query failed, %s", err.Error())
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(err.Error())
		}

		if nextData != nil {
			// queued阶段只更新nextUri，不处理数据
			nextUri = nextData.NextUri
		}
	}

	// 第二阶段：处理executing阶段
	return fs.getDataFromEtrinoExecutingNextUri(ctx, nextUri, queryType, queryId, slug)
}

// getDataFromEtrinoExecutingNextUri 从Etrino获取executing阶段的数据
func (fs *Service) getDataFromEtrinoExecutingNextUri(ctx context.Context, nextUri string, queryType int, queryId string, slug string) error {

	// 初始化结果通道缓存
	queryCacheKey := fmt.Sprintf("%s_%s", queryId, slug)

	existingCache, ok := fs.queryCache.Load(queryCacheKey)
	if !ok {
		logger.Errorf("Query does not exist, queryId: %s, slug: %s", queryId, slug)
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails("Query does not exist")
	}
	resultCache, _ := existingCache.(*interfaces.ResultCache)

	resultCache.ResultSet = nextUri

	// 提交查询任务到查询池
	err := fs.queryPool.Submit(func() {

		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("QueryId: %s, slug: %s, etrino query panic in goroutine: %v", queryId, slug, r)
				logger.Error(err)
				fs.storeToCache(resultCache, nil, nil, nil, err)
			}

			logger.Debugf("QueryId: %s, slug: %s, etrino query goroutine exit", queryId, slug)
		}()

		var columns []*interfaces.Column
		var queryErr error

		for {
			var data []*[]any

			for nextUri != "" {
				// 如果缓存不存在, 则退出循环
				if _, ok := fs.queryCache.Load(queryCacheKey); !ok {
					goto CacheNotFound
				}

				if queryType == 2 {
					nextUri = fmt.Sprintf("%s?targetResultBatchSize=%d", nextUri, fs.appSetting.QuerySetting.DataQuerySize)
				}

				nextData, err := fs.vegaCalculateAccess.NextUriQuery(ctx, nextUri)

				if err != nil {
					nextUri = ""
					queryErr = err
					break
				} else {
					nextUri = nextData.NextUri
					columns = nextData.Columns

					if nextData.Data != nil {
						data = append(data, nextData.Data...)

						if nextUri != "" && queryType == 2 {
							break
						}
					}
				}
			}

			// 缓存结果
			if nextUri != "" {
				fs.storeToCache(resultCache, nextUri, columns, data, queryErr)
			} else {
				fs.storeToCache(resultCache, nil, columns, data, queryErr)
			}

			// 发生错误，退出循环
			if queryErr != nil {
				break
			}

			// 没有更多数据，退出循环
			if nextUri == "" {
				break
			}

			// 定期检查缓存中是否有数据
			fs.checkCachePeriodically(resultCache)
		}
	CacheNotFound:
	})

	if err != nil {
		logger.Errorf("Get data from Etrino executing nextUri failed, %s", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(err.Error())
	}

	return nil
}

// handleQueryResult 处理查询结果，根据查询类型和超时时间返回结果
// 如果是同步查询，等待所有数据处理完成后返回
// 如果是流式查询，根据超时时间返回结果或nextUri
func (fs *Service) handleQueryResult(ctx context.Context, queryType int, timeout int, batchSize int, queryId string, slug string, token int) (*interfaces.FetchResp, error) {

	queryCacheKey := fmt.Sprintf("%s_%s", queryId, slug)
	existingCache, ok := fs.queryCache.Load(queryCacheKey)
	if !ok {
		logger.Errorf("Query does not exist, queryId: %s, slug: %s, token: %d", queryId, slug, token)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails("Query does not exist")
	}
	resultCache, _ := existingCache.(*interfaces.ResultCache)

	if queryType == 2 && timeout > 0 {
		// 如果是流式查询且存在超时时间时，判断超时时间内是否有数据

		// 创建定期和超时通道
		timeoutChan := time.After(time.Duration(timeout) * time.Second)
		ticker := time.NewTicker(1000 * time.Millisecond) // 每1000毫秒检查一次缓存
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// 定期检查缓存， 查询异常、数据足够、查询完成退出检查
				if resultCache.Error != nil ||
					len(resultCache.ResultChan) >= batchSize ||
					(resultCache.Columns != nil && resultCache.ResultSet == nil) {
					break
				}
				continue
			case <-timeoutChan:
				// 超时后检查缓存， 查询异常、数据足够、查询完成退出检查
				if resultCache.Error != nil ||
					len(resultCache.ResultChan) >= batchSize ||
					(resultCache.Columns != nil && resultCache.ResultSet == nil) {
					break
				} else {
					resultCache.Token++
					nextUri := fmt.Sprintf("http://%s:%d/api/vega-gateway/v2/fetch/%s/%s/%d",
						version.ServerName, fs.appSetting.ServerSetting.HttpPort, queryId, slug, token+1)
					finalResult := &interfaces.FetchResp{
						NextUri: nextUri,
					}
					return finalResult, nil
				}
			default:
				continue
			}
			break
		}
	}

	data := make([]*[]any, 0)
	var totalCount int

	// 从缓存中获取查询结果，直到数据足够或查询完成
	for {
		if len(resultCache.ResultChan) > 0 {
			for {
				chanData := <-resultCache.ResultChan
				data = append(data, chanData)
				totalCount++
				if queryType == 2 && totalCount == batchSize {
					break
				}
				if len(resultCache.ResultChan) > 0 {
					continue
				}
				break
			}
		}

		if (queryType == 2 && totalCount == batchSize) ||
			(resultCache.ResultSet == nil && len(resultCache.ResultChan) == 0) {
			break
		}
	}

	if resultCache.Error != nil {
		logger.Errorf("Query failed, %s", resultCache.Error.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(resultCache.Error.Error())
	}

	// 构建最终结果
	finalResult := &interfaces.FetchResp{
		Columns:    resultCache.Columns,
		Entries:    data,
		TotalCount: int64(totalCount),
	}

	// 检查是否还有更多数据
	if resultCache.ResultSet != nil || len(resultCache.ResultChan) > 0 {
		resultCache.Token++
		finalResult.NextUri = fmt.Sprintf("http://%s:%d/api/vega-gateway/v2/fetch/%s/%s/%d",
			version.ServerName, fs.appSetting.ServerSetting.HttpPort, queryId, slug, resultCache.Token)
	}

	return finalResult, nil
}

func (fs *Service) NextQuery(ctx context.Context, req *interfaces.NextQueryReq) (resp *interfaces.FetchResp, err error) {

	startTime := time.Now()

	// 构建缓存键
	queryCacheKey := fmt.Sprintf("%s_%s", req.QueryId, req.Slug)

	// 获取或创建该查询的锁
	lock, _ := fs.queryLocks.LoadOrStore(queryCacheKey, &sync.Mutex{})
	queryLock := lock.(*sync.Mutex)

	// 加锁
	queryLock.Lock()
	defer queryLock.Unlock()

	// 从缓存中获取查询结果
	cachedResult, ok := fs.queryCache.Load(queryCacheKey)
	if !ok {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, rest.PublicError_NotFound).
			WithErrorDetails("Query does not exist")
	}
	resultCache, _ := cachedResult.(*interfaces.ResultCache)

	if resultCache.Token != req.Token {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("Token not match")
	}

	defer func() {

		if r := recover(); r != nil { // 处理查询中的panic
			logger.Errorf("QueryId: %s, slug: %s, token: %d, query panic in goroutine: %v", req.QueryId, req.Slug, req.Token, r)
			fs.cleanQuery(queryCacheKey, resultCache)
		} else if ctx.Err() != nil { // 处理查询上下文取消
			logger.Errorf("QueryId: %s, slug: %s, token: %d, context err: %v", req.QueryId, req.Slug, req.Token, ctx.Err())
			fs.cleanQuery(queryCacheKey, resultCache)
		} else if err != nil { // 处理查询错误
			logger.Errorf("QueryId: %s, slug: %s, token: %d, next query failed with error: %v", req.QueryId, req.Slug, req.Token, err)
			fs.cleanQuery(queryCacheKey, resultCache)
		} else if resultCache.ResultSet == nil && len(resultCache.ResultChan) == 0 { // 处理查询完成
			logger.Debugf("QueryId: %s, slug: %s, token: %d, query completed", req.QueryId, req.Slug, req.Token)
			fs.cleanQuery(queryCacheKey, resultCache)
		}

		// 记录执行耗时
		logger.Debugf("Next query in %v, queryId: %s, slug: %s, token: %d", time.Since(startTime), req.QueryId, req.Slug, req.Token)
	}()

	return fs.handleQueryResult(ctx, 2, 0, req.BatchSize, req.QueryId, req.Slug, req.Token)

}

// generateQueryId 生成格式为 YYYMMdd_HHmmss_index_coordId 的查询ID
func (fs *Service) generateQueryId() string {
	now := time.Now()
	nowSec := now.Unix()
	nowDay := now.Unix() / 86400 // 计算天数

	// 计数器超过99999时等待秒数变化
	if fs.queryIdCounter > 99999 {
		for now.Unix() == fs.lastTimeInSec {
			time.Sleep(time.Second)
			now = time.Now()
		}
		fs.queryIdCounter = 0
	}

	// 如果秒数变化，更新时间戳
	if nowSec != fs.lastTimeInSec {
		fs.lastTimeInSec = nowSec
		fs.lastTimestamp = now.Format("20060102_150405")

		// 如果天数变化，重置计数器
		if nowDay != fs.lastTimeInDay {
			fs.lastTimeInDay = nowDay
			fs.queryIdCounter = 0
		}
	}

	// 生成ID并递增计数器
	id := fmt.Sprintf("%s_%05d_%s", fs.lastTimestamp, fs.queryIdCounter, fs.coordinatorId)
	fs.queryIdCounter++
	return id
}

// generateRandomCoordId 生成5位随机字符串作为coordinatorId
func generateRandomCoordId() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 5)
	for i := range b {
		b[i] = letters[mathRand.Intn(len(letters))]
	}
	return string(b)
}

func generateSlug() string {
	// 生成UUID
	uuid := make([]byte, 16)
	_, err := cryptoRand.Read(uuid)
	if err != nil {
		panic(fmt.Sprintf("生成UUID失败: %v", err))
	}

	// 设置UUID版本和变体
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	// 格式化为字符串并移除连字符
	uuidStr := fmt.Sprintf("%x%x%x%x%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])

	// 添加"x"前缀并返回
	return "x" + strings.ToLower(uuidStr)
}
