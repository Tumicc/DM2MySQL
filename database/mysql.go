package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLConnector 封装 MySQL 连接操作
type MySQLConnector struct {
	db      *sql.DB
	version int // 例如: 5 代表 MySQL 5.7, 8 代表 MySQL 8.0+
}

// MySQLColumn 定义 MySQL 列元数据结构
type MySQLColumn struct {
	Name          string
	DataType      string // 数据库原始类型字符串 (如 "NUMBER", "VARCHAR2")
	DataLength    int64  // 字节长度
	DataPrecision int    // 数字总位数
	DataScale     int    // 小数位数
	Nullable      bool   // true 表示可为空, false 表示必填
	IsPrimaryKey  bool   // 是否为主键
	IsAutoIncrement bool // 是否为自增列
}

// NewMySQLConnector 初始化 MySQL 连接
func NewMySQLConnector(dsn string, version int) (*MySQLConnector, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 优化连接池配置，适应生产环境数据迁移场景
	db.SetMaxOpenConns(20)                  // 最大打开连接数
	db.SetMaxIdleConns(10)                  // 最大空闲连接数
	db.SetConnMaxLifetime(10 * time.Minute) // 连接最大存活时间

	// 尝试 Ping 一下确保连接成功
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &MySQLConnector{db: db, version: version}, nil
}

// Close 关闭数据库连接
func (mc *MySQLConnector) Close() error {
	return mc.db.Close()
}

// DisableConstraints 禁用外键和唯一约束检查（用于数据导入前）
func (mc *MySQLConnector) DisableConstraints() error {
	// 允许在一个 Exec 中执行多条语句需 DSN开启 multiStatements=true，
	// 为保险起见，这里分开执行
	_, err := mc.db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	if err != nil {
		return err
	}
	_, err = mc.db.Exec("SET UNIQUE_CHECKS = 0")
	return err
}

// EnableConstraints 启用外键和唯一约束检查（数据导入后恢复）
func (mc *MySQLConnector) EnableConstraints() error {
	_, err := mc.db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	if err != nil {
		return err
	}
	_, err = mc.db.Exec("SET UNIQUE_CHECKS = 1")
	return err
}

// CreateTable 根据通用的 Column 定义创建 MySQL 表
func (mc *MySQLConnector) CreateTable(tableName string, columns []MySQLColumn) error {
	// 检查是否有列定义
	if len(columns) == 0 {
		return fmt.Errorf("表 %s 没有列定义，无法创建表", tableName)
	}
	
	// 1. 删除旧表
	_, err := mc.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName))
	if err != nil {
		return fmt.Errorf("drop table error: %v", err)
	}

	// 2. 构建字段定义
	var colDefs []string
	var primaryKeys []string // 收集主键列
	
	for _, col := range columns {
		// 获取映射后的 MySQL 类型
		colType := convertDMTypeToMySQL(col, mc.version)

		// 处理 Nullable 属性
		nullDef := "NULL"
		if !col.Nullable {
			nullDef = "NOT NULL"
		}

		// 组装: `字段名` 类型 NULL/NOT NULL
		def := fmt.Sprintf("`%s` %s %s", col.Name, colType, nullDef)
		
		// 如果是自增列
		if col.IsAutoIncrement {
			def += " AUTO_INCREMENT"
		}
		
		// 收集主键列
		if col.IsPrimaryKey {
			primaryKeys = append(primaryKeys, "`" + col.Name + "`")
		}
		
		colDefs = append(colDefs, def)
	}
	
	// 如果有主键，添加主键约束
	if len(primaryKeys) > 0 {
		pkConstraint := fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", "))
		colDefs = append(colDefs, pkConstraint)
	}

	// 3. 组装 CREATE TABLE 语句
	sqlStr := fmt.Sprintf("CREATE TABLE `%s` (%s) ENGINE=InnoDB", tableName, strings.Join(colDefs, ","))

	// 4. 根据版本追加字符集设置
	if mc.version >= 8 {
		// MySQL 8.0+: 推荐 utf8mb4 和 DYNAMIC 行格式
		sqlStr += " DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC"
	} else {
		// MySQL 5.7: 使用 utf8 字符集（兼容旧版）
		sqlStr += " DEFAULT CHARSET=utf8"
	}

	_, err = mc.db.Exec(sqlStr)
	if err != nil {
		return fmt.Errorf("create table error: %v, sql: %s", err, sqlStr)
	}
	return nil
}

// convertDMTypeToMySQL 将达梦/Oracle 类型映射为最佳的 MySQL 类型
func convertDMTypeToMySQL(col MySQLColumn, version int) string {
	// 转大写并去除首尾空格，防止 " INT " 这种奇怪情况
	originType := strings.ToUpper(strings.TrimSpace(col.DataType))

	// --- 1. 优先匹配标准 SQL 整数类型 (修复 INT/BIGINT 变 TEXT 的问题) ---
	switch originType {
	case "BIGINT":
		return "BIGINT"
	case "INT", "INTEGER":
		return "INT"
	case "SMALLINT":
		return "SMALLINT"
	case "TINYINT", "BYTE":
		return "TINYINT"
	case "BIT", "BOOL", "BOOLEAN":
		// MySQL BIT 类型操作不便，业界通常用 TINYINT(1) 代替
		return "TINYINT(1)"
	case "REAL", "DOUBLE", "DOUBLE PRECISION", "FLOAT":
		return "DOUBLE"
	}

	// --- 2. 处理 Oracle/达梦 风格的数值类型 (NUMBER/DECIMAL) ---
	if strings.Contains(originType, "NUMBER") ||
		strings.Contains(originType, "DECIMAL") ||
		strings.Contains(originType, "NUMERIC") ||
		strings.Contains(originType, "DEC") {

		p := col.DataPrecision
		s := col.DataScale

		// 如果未指定精度 (如 NUMBER)，默认为最大精度 DECIMAL
		if p == 0 && s == 0 {
			return "DECIMAL(38,4)"
		}

		// 如果 Scale 为 0，说明是纯整数，尝试转换为更高效的整型
		if s == 0 {
			switch {
			case p <= 3: // -128 ~ 127
				return "TINYINT"
			case p <= 5: // -32768 ~ 32767
				return "SMALLINT"
			case p <= 9: // -21亿 ~ 21亿 (int32)
				return "INT"
			case p <= 19: // int64 范围
				return "BIGINT"
			default:
				// 超过 19 位，BIGINT 存不下，必须用 DECIMAL
				return fmt.Sprintf("DECIMAL(%d,0)", p)
			}
		}

		// 如果有小数位，使用 DECIMAL
		// MySQL 限制: Precision <= 65, Scale <= 30, 且 Scale <= Precision
		if p > 65 {
			p = 65
		}
		if s > 30 {
			s = 30
		}
		if p < s {
			p = s
		}
		return fmt.Sprintf("DECIMAL(%d,%d)", p, s)
	}

	// --- 3. 处理字符串类型 ---
	if strings.Contains(originType, "CHAR") || strings.Contains(originType, "STR") {
		length := col.DataLength

		// 安全阈值判断：MySQL 单行最大约 65535 字节
		// utf8 下，1字符=3字节。如果定义太长，转为 TEXT/LONGTEXT 以避免报错
		if length > 21845 { // 65535/3 = 21845
			return "LONGTEXT"
		} else if length > 5461 { // 16383/3 = 5461
			// 5461 字节以上通常建议用 TEXT，避免占用行缓冲
			return "TEXT"
		} else {
			return fmt.Sprintf("VARCHAR(%d)", length)
		}
	}

	// --- 4. 处理时间日期 ---
	// 达梦 DATE 含时间，对应 MySQL DATETIME
	if originType == "DATE" {
		return "DATETIME"
	}
	if strings.Contains(originType, "TIME") {
		// TIMESTAMP 映射为 DATETIME，对于 MySQL 5.x 不支持 (6) 精度
		if strings.Contains(originType, "TIMESTAMP") {
			if version >= 8 {
				return "DATETIME(6)" // MySQL 8.0 支持微秒精度
			}
			return "DATETIME" // MySQL 5.x 不支持微秒精度
		}
		return "DATETIME"
	}

	// --- 5. 处理 LOB (大对象) ---
	if strings.Contains(originType, "CLOB") || strings.Contains(originType, "TEXT") || strings.Contains(originType, "LONGVARCHAR") {
		return "LONGTEXT"
	}
	if strings.Contains(originType, "BLOB") || strings.Contains(originType, "IMAGE") || strings.Contains(originType, "BINARY") {
		return "LONGBLOB"
	}

	// --- 6. 兜底 ---
	return "LONGTEXT"
}

// execWithRetry 带重试机制的执行函数
func (mc *MySQLConnector) execWithRetry(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error
	
	// 最多重试3次
	for i := 0; i < 3; i++ {
		result, err = mc.db.ExecContext(ctx, query, args...)
		if err == nil {
			return result, nil
		}
		
		// 如果是连接错误，尝试重新连接
		if strings.Contains(err.Error(), "connection refused") || 
		   strings.Contains(err.Error(), "broken pipe") || 
		   strings.Contains(err.Error(), "invalid connection") ||
		   strings.Contains(err.Error(), "connection lost") {
			log.Printf("⚠️  检测到连接问题，尝试重新连接 (%d/3)", i+1)
			time.Sleep(time.Duration(i+1) * time.Second) // 逐渐增加等待时间
			continue
		}
		
		// 非连接错误直接返回
		break
	}
	
	return result, err
}

// BatchInsertData 执行分批插入
// rows: 源数据库查询结果集
// userBatchSize: 用户期望的每批次行数 (会自动调整以适应 MySQL 占位符限制)
func (mc *MySQLConnector) BatchInsertData(tableName string, columns []MySQLColumn, rows *sql.Rows, userBatchSize int) (int64, error) {
	colCount := len(columns)
	if colCount == 0 {
		return 0, nil
	}

	// 计算安全的 batchSize
	// MySQL 预处理语句参数限制通常为 65535，保险起见设为 60000
	maxPlaceholders := 60000
	safeBatchSize := maxPlaceholders / colCount

	// 使用较小的值
	if safeBatchSize < userBatchSize {
		userBatchSize = safeBatchSize
	}
	if userBatchSize < 1 {
		userBatchSize = 1
	}

	log.Printf("📝 表 %s 批处理大小设置为 %d (每批 %d 行, %d 列)", tableName, userBatchSize, userBatchSize, colCount)

	// 准备 SQL 模板
	// 结果如: INSERT INTO `table` (`col1`, `col2`) VALUES
	colNames := make([]string, colCount)
	placeholders := make([]string, colCount)
	for i, col := range columns {
		colNames[i] = "`" + col.Name + "`"
		placeholders[i] = "?"
	}
	baseSQL := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES ", tableName, strings.Join(colNames, ", "))
	rowPlaceholder := fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))

	// 变量初始化
	var totalRows int64 = 0
	var batchValues []interface{}
	var batchPlaceholders []string

	// 用于 Scan 的容器
	scanArgs := make([]interface{}, colCount)
	values := make([]interface{}, colCount)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// 遍历数据
	lastReportTime := time.Now()
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			return totalRows, fmt.Errorf("scan rows error: %v", err)
		}

		// 处理读取到的数据
		for _, v := range values {
			if b, ok := v.([]byte); ok {
				// 将 []byte 转为 string，防止某些情况下的乱码或 hex 显示
				batchValues = append(batchValues, string(b))
			} else {
				batchValues = append(batchValues, v)
			}
		}

		batchPlaceholders = append(batchPlaceholders, rowPlaceholder)
		totalRows++

		// 缓冲区满，执行插入
		if len(batchPlaceholders) >= userBatchSize {
			stmt := baseSQL + strings.Join(batchPlaceholders, ",")
			log.Printf("📤 正在插入表 %s 的一批数据 (%d 行)", tableName, len(batchPlaceholders))
			
			// 创建带超时的上下文
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			_, err := mc.execWithRetry(ctx, stmt, batchValues...)
			cancel()
			
			if err != nil {
				return totalRows, fmt.Errorf("batch exec error: %v", err)
			}
			
			log.Printf("📥 表 %s 的一批数据插入完成 (%d 行)", tableName, len(batchPlaceholders))
			
			// 清空缓冲区
			batchValues = nil
			batchPlaceholders = nil
			
			// 每隔一段时间报告一次进度
			if time.Since(lastReportTime) > 30*time.Second {
				log.Printf("📊 表 %s 已处理 %d 行", tableName, totalRows)
				lastReportTime = time.Now()
			}
		}
	}

	// 处理剩余数据
	if len(batchPlaceholders) > 0 {
		stmt := baseSQL + strings.Join(batchPlaceholders, ",")
		log.Printf("📤 正在插入表 %s 的最后一批数据 (%d 行)", tableName, len(batchPlaceholders))
		
		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		_, err := mc.execWithRetry(ctx, stmt, batchValues...)
		cancel()
		
		if err != nil {
			return totalRows, fmt.Errorf("final batch exec error: %v", err)
		}
		
		log.Printf("📥 表 %s 的最后一批数据插入完成 (%d 行)", tableName, len(batchPlaceholders))
	}

	return totalRows, nil
}
