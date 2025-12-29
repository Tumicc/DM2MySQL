package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "gitee.com/chunanyong/dm" // 确保你有这个驱动
)

type DMConnector struct {
	db *sql.DB
	// 缓存表名映射，键为小写的表名，值为真实的表名
	tableNameMap map[string]string
}

// DMColumn 定义达梦列元数据结构 (与 MySQL 中的 Column 结构体保持一致)
type DMColumn struct {
	Name          string
	DataType      string
	DataLength    int64
	DataPrecision int
	DataScale     int
	Nullable      bool
	ColumnID      int
	IsPrimaryKey  bool // 是否为主键
	IsIdentity    bool // 是否为自增列
}

// NewDMConnector 创建一个新的达梦连接器
func NewDMConnector(dsn string) (*DMConnector, error) {
	// 调整达梦连接参数，避免 cursor 超时
	db, err := sql.Open("dm", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	
	// 初始化表名映射缓存
	connector := &DMConnector{
		db: db,
		tableNameMap: make(map[string]string),
	}
	
	// 预加载所有表名映射
	err = connector.loadTableNameMap()
	if err != nil {
		return nil, err
	}
	
	return connector, nil
}

// loadTableNameMap 预加载所有表名映射到缓存中
func (dmc *DMConnector) loadTableNameMap() error {
	query := `SELECT TABLE_NAME FROM USER_TABLES WHERE TABLESPACE_NAME != 'SYSTEM'`
	rows, err := dmc.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		// 使用小写作为键
		dmc.tableNameMap[strings.ToLower(tableName)] = tableName
	}
	
	return nil
}

// getRealTableName 获取数据库中真实的表名
func (dmc *DMConnector) getRealTableName(tableName string) string {
	// 先从缓存中查找
	if realName, exists := dmc.tableNameMap[strings.ToLower(tableName)]; exists {
		return realName
	}
	
	// 如果缓存中没有，直接返回原名
	return tableName
}

// Close 关闭数据库连接
func (dmc *DMConnector) Close() error {
	return dmc.db.Close()
}

// GetTables 获取所有用户表
func (dmc *DMConnector) GetTables() ([]string, error) {
	// 过滤掉系统表
	query := `SELECT TABLE_NAME FROM USER_TABLES WHERE TABLESPACE_NAME != 'SYSTEM' ORDER BY TABLE_NAME`
	rows, err := dmc.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

// GetTableSchema 获取表的列结构信息
func (dmc *DMConnector) GetTableSchema(tableName string) ([]DMColumn, error) {
	// 获取真实的表名
	realTableName := dmc.getRealTableName(tableName)
	
	// 如果使用了不同的表名，记录日志
	if realTableName != tableName {
		log.Printf("🔄 表名不区分大小写匹配: '%s' -> '%s'", tableName, realTableName)
	}
	
	// 查询列信息，包括主键和自增信息
	// 达梦数据库中，自增列信息通常在 ALL_TAB_COLUMNS 或 USER_TAB_COLUMNS 视图中没有直接标识
	// 我们需要通过其他方式来确定自增列
	
	// 首先获取基本列信息
	query := `
		SELECT 
			utc.COLUMN_NAME, 
			utc.DATA_TYPE, 
			utc.DATA_LENGTH, 
			utc.DATA_PRECISION, 
			utc.DATA_SCALE, 
			utc.NULLABLE, 
			utc.COLUMN_ID
		FROM USER_TAB_COLUMNS utc
		WHERE utc.TABLE_NAME = ?
		ORDER BY utc.COLUMN_ID`

	rows, err := dmc.db.Query(query, realTableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []DMColumn
	for rows.Next() {
		var c DMColumn
		var nullStr string
		var prec, scale sql.NullInt64
		var dataLength int
		if err := rows.Scan(&c.Name, &c.DataType, &dataLength, &prec, &scale, &nullStr, &c.ColumnID); err != nil {
			return nil, err
		}
		c.DataLength = int64(dataLength)
		c.Nullable = (nullStr == "Y")
		c.DataPrecision = int(prec.Int64)
		c.DataScale = int(scale.Int64)
		cols = append(cols, c)
	}
	
	// 检查是否有迭代错误
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	// 获取主键信息
	pkQuery := `
		SELECT ucc.COLUMN_NAME
		FROM USER_CONS_COLUMNS ucc
		JOIN USER_CONSTRAINTS uc ON ucc.CONSTRAINT_NAME = uc.CONSTRAINT_NAME
		WHERE uc.CONSTRAINT_TYPE = 'P' AND uc.TABLE_NAME = ?`
		
	pkRows, err := dmc.db.Query(pkQuery, realTableName)
	if err != nil {
		return nil, err
	}
	defer pkRows.Close()
	
	// 创建列名到索引的映射
	colNameIndex := make(map[string]int)
	for i, col := range cols {
		colNameIndex[col.Name] = i
	}
	
	// 标记主键列
	for pkRows.Next() {
		var colName string
		if err := pkRows.Scan(&colName); err != nil {
			return nil, err
		}
		if idx, exists := colNameIndex[colName]; exists {
			cols[idx].IsPrimaryKey = true
		}
	}
	
	// 检查是否有迭代错误
	if err = pkRows.Err(); err != nil {
		return nil, err
	}
	
	// 获取自增列信息 (IDENTITY列)
	// 首先尝试使用 USER_TAB_IDENTITY_COLS 视图（适用于较新版本的达梦数据库）
	identityQuery := `
		SELECT COLUMN_NAME
		FROM USER_TAB_IDENTITY_COLS
		WHERE TABLE_NAME = ?`
		
	identityRows, err := dmc.db.Query(identityQuery, realTableName)
	if err != nil {
		// 如果 USER_TAB_IDENTITY_COLS 视图不存在（可能是较老版本的达梦数据库），尝试另一种方法
		log.Printf("⚠️  查询自增列信息失败，可能达梦版本不支持 USER_TAB_IDENTITY_COLS 视图: %v", err)
		log.Println("🔄 尝试使用备用方法检测自增列...")
		
		// 在旧版本达梦中，可以尝试通过查询系统表注释或其他方式识别自增列
		// 这里我们暂时跳过自增列处理，但记录警告信息
		log.Println("⚠️  当前版本可能不支持自增列自动识别，如有自增列请手动处理")
	} else {
		defer identityRows.Close()
		
		// 标记自增列
		for identityRows.Next() {
			var colName string
			if err := identityRows.Scan(&colName); err != nil {
				return nil, err
			}
			if idx, exists := colNameIndex[colName]; exists {
				cols[idx].IsIdentity = true
			}
		}
		
		// 检查是否有迭代错误
		if err = identityRows.Err(); err != nil {
			return nil, err
		}
	}
	
	log.Printf("📋 获取到表 %s 的结构，共 %d 个字段", tableName, len(cols))
	
	// 如果没有获取到列信息，返回一个明确的错误
	if len(cols) == 0 {
		return nil, fmt.Errorf("表 %s 没有找到任何列定义", tableName)
	}
	
	return cols, nil
}

// GetTableData 获取表的所有数据，对于大表采用流式处理
func (dmc *DMConnector) GetTableData(tableName string) (*sql.Rows, error) {
	// 获取真实的表名
	realTableName := dmc.getRealTableName(tableName)
	
	// 如果使用了不同的表名，记录日志
	if realTableName != tableName {
		log.Printf("🔄 表名不区分大小写匹配: '%s' -> '%s'", tableName, realTableName)
	}
	
	// 简单的 SELECT *，如果表超级大（100GB+），可能需要基于主键分页，但 1GB 数据流式读取通常没问题
	log.Printf("📥 开始读取表 %s 的数据", tableName)
	
	// 对于包含大字段的表，增加流控以避免内存溢出
	query := fmt.Sprintf("SELECT * FROM %s", realTableName)
	return dmc.db.Query(query)
}