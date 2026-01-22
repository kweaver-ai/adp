package sql_connectors

import (
	"context"
	"database/sql"
	"fmt"
	"vega-gateway-pro/interfaces"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

// MySQLConnector MySQL连接器
type MySQLConnector struct {
	ConnInfo *interfaces.DataSource
	db       *sql.DB
}

// NewMySQLConnector 创建MySQL连接器
func NewMySQLConnector(connInfo *interfaces.DataSource) (*MySQLConnector, error) {

	newMySQLConnector := &MySQLConnector{
		ConnInfo: connInfo,
	}

	// 1. 构建DSN连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&loc=Local",
		connInfo.BinData.Account,
		connInfo.BinData.Password,
		connInfo.BinData.Host,
		connInfo.BinData.Port,
		connInfo.BinData.DataBaseName)

	// 2. 打开数据库连接
	var err error
	newMySQLConnector.db, err = sql.Open("mysql", dsn)
	if err != nil {
		logger.Errorf("connect mysql failed: %v", err)
		return nil, fmt.Errorf("connect mysql failed: %w", err)
	}

	// 3. 验证连接
	if err := newMySQLConnector.db.PingContext(context.Background()); err != nil {
		newMySQLConnector.db.Close()
		logger.Errorf("connect mysql failed: %v", err)
		return nil, fmt.Errorf("connect mysql failed: %w", err)
	}

	return newMySQLConnector, nil
}

// GetResultSet 获取查询结果集
func (h *MySQLConnector) GetResultSet(query string) (any, error) {

	rows, err := h.db.QueryContext(context.Background(), query)
	if err != nil {
		h.db.Close()
		logger.Errorf("query failed: %v", err)
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return rows, nil
}

// GetColumns 获取结果集中的列信息
func (h *MySQLConnector) GetColumns(resultSet any) ([]*interfaces.Column, error) {

	rows, ok := resultSet.(*sql.Rows)
	if !ok {
		logger.Errorf("resultSet type error, expect *sql.Rows")
		return nil, fmt.Errorf("resultSet type error, expect *sql.Rows")
	}

	// 获取列名和列类型
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		rows.Close()
		h.db.Close()
		logger.Errorf("get column types failed: %v", err)
		return nil, fmt.Errorf("get column types failed: %w", err)
	}

	// 收集列信息
	var columns []*interfaces.Column
	for _, ct := range columnTypes {
		columns = append(columns, &interfaces.Column{
			Name: ct.Name(),
			Type: ct.DatabaseTypeName(),
		})
	}

	return columns, nil
}

// GetData 从结果集中提取数据
func (h *MySQLConnector) GetData(resultSet any, columnSize int, batchSize int) (any, []*[]any, error) {
	rows, ok := resultSet.(*sql.Rows)
	if !ok {
		logger.Errorf("resultSet type error, expect *sql.Rows")
		return nil, nil, fmt.Errorf("resultSet type error, expect *sql.Rows")
	}

	var result []*[]interface{}
	values := make([]interface{}, columnSize)
	valuePtrs := make([]interface{}, columnSize)
	// 遍历结果行
	for rows.Next() {
		// 初始化指针切片
		for i := 0; i < columnSize; i++ {
			valuePtrs[i] = &values[i]
		}

		// 扫描行数据
		if err := rows.Scan(valuePtrs...); err != nil {
			rows.Close()
			h.db.Close()
			logger.Errorf("scan row data failed: %v", err)
			return nil, nil, fmt.Errorf("scan row data failed: %w", err)
		}

		// 转换为数组格式
		row := make([]interface{}, columnSize)
		for i, val := range values {
			if b, ok := val.([]byte); ok {
				row[i] = string(b)
			} else {
				row[i] = val
			}
		}
		result = append(result, &row)
		// 批次处理
		if len(result) >= batchSize {
			break
		}
	}

	// 检查行遍历过程中的错误
	if err := rows.Err(); err != nil {
		rows.Close()
		h.db.Close()
		logger.Errorf("scan rows failed: %v", err)
		return nil, nil, fmt.Errorf("scan rows failed: %w", err)
	}

	// 判断是否还有更多数据
	if len(result) < batchSize {
		// 如果结果数量小于批次大小，说明没有更多数据了
		rows.Close()
		h.db.Close()
		return nil, result, nil
	} else {
		return rows, result, nil
	}
}

func (h *MySQLConnector) Close() error {
	return h.db.Close()
}
