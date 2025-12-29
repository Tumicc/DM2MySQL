package main

import (
	"context"
	"dm2mysql-migrator/config"
	"dm2mysql-migrator/database"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	// --- 达梦参数 ---
	dmHost   = flag.String("dm-host", "127.0.0.1", "达梦数据库IP")
	dmPort   = flag.Int("dm-port", 5236, "达梦数据库端口")
	dmUser   = flag.String("dm-user", "", "达梦用户名")
	dmPass   = flag.String("dm-pass", "", "达梦密码")
	dmSchema = flag.String("dm-schema", "", "达梦模式名")
	dmExtra  = flag.String("dm-extra", "", "达梦额外参数")

	// --- MySQL 参数 ---
	mysqlHost  = flag.String("mysql-host", "127.0.0.1", "MySQL数据库IP")
	mysqlPort  = flag.Int("mysql-port", 3306, "MySQL数据库端口")
	mysqlUser  = flag.String("mysql-user", "root", "MySQL用户名")
	mysqlPass  = flag.String("mysql-pass", "", "MySQL密码")
	mysqlDB    = flag.String("mysql-db", "", "MySQL数据库名")
	mysqlVer   = flag.Int("mysql-ver", 5, "MySQL版本: 5 (代表5.0-5.7) 或 8 (代表8.0+)") // 新增参数
	mysqlExtra = flag.String("mysql-extra", "", "MySQL额外参数")

	// --- 全局 ---
	workerNum = flag.Int("workers", 4, "并发数")
	batchSize = flag.Int("batch", 2000, "批量大小")

	// --- 配置文件 ---
	tablesConfigFile = flag.String("tables-config", "./config/tables.json", "表配置文件路径")
)

func buildDMDSN() string {
	dsn := fmt.Sprintf("dm://%s:%s@%s:%d", *dmUser, *dmPass, *dmHost, *dmPort)
	var params []string
	if *dmSchema != "" {
		params = append(params, "schema="+*dmSchema)
	}
	if *dmExtra != "" {
		params = append(params, *dmExtra)
	}
	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}
	return dsn
}

func buildMySQLDSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", *mysqlUser, *mysqlPass, *mysqlHost, *mysqlPort, *mysqlDB)

	params := []string{
		"parseTime=True",
		"loc=Local",
		"interpolateParams=true",
		"timeout=30s",
		"readTimeout=30s",
		"writeTimeout=30s",
	}

	// 核心区别 1: 字符集选择
	if *mysqlVer >= 8 {
		params = append(params, "charset=utf8mb4") // 8.0+ 使用 utf8mb4
		log.Println("💡 检测到 MySQL 8.0+ 模式: 使用 utf8mb4 字符集")
	} else {
		params = append(params, "charset=utf8") // 5.0 使用 utf8 (3字节)
		log.Println("💡 检测到 MySQL 5.x 模式: 使用 utf8 字符集 (兼容旧版)")
	}

	if *mysqlExtra != "" {
		params = append(params, *mysqlExtra)
	}
	dsn += "?" + strings.Join(params, "&")
	return dsn
}

func main() {
	flag.Parse()

	// 校验
	if *dmUser == "" || *dmPass == "" || *dmSchema == "" {
		fmt.Println("❌ 达梦参数缺失")
		flag.Usage()
		os.Exit(1)
	}
	if *mysqlUser == "" || *mysqlPass == "" || *mysqlDB == "" {
		fmt.Println("❌ MySQL参数缺失")
		flag.Usage()
		os.Exit(1)
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🚀 开始数据库迁移...")

	// 加载表配置
	tablesConfig, err := config.LoadTablesConfig(*tablesConfigFile)
	if err != nil {
		log.Fatalf("加载表配置文件失败: %v", err)
	}

	tables := tablesConfig.Tables

	// 初始化
	startTime := time.Now()
	log.Println("🔗 正在连接到达梦数据库...")
	dmConn, err := database.NewDMConnector(buildDMDSN())
	if err != nil {
		log.Fatalf("达梦连接失败: %v", err)
	}
	defer dmConn.Close()
	log.Println("✅ 达梦数据库连接成功")

	log.Println("🔗 正在连接到MySQL数据库...")
	// 传入版本号到 Connector
	mysqlConn, err := database.NewMySQLConnector(buildMySQLDSN(), *mysqlVer)
	if err != nil {
		log.Fatalf("MySQL连接失败: %v", err)
	}
	defer mysqlConn.Close()
	log.Println("✅ MySQL数据库连接成功")

	// 准备
	log.Println("⚙️  正在禁用约束检查...")
	mysqlConn.DisableConstraints()
	log.Println("✅ 约束检查已禁用")

	log.Printf("📋 准备迁移指定的 %d 张表...", len(tables))

	// 并发
	var wg sync.WaitGroup
	jobs := make(chan string, len(tables))

	// 创建一个map来存储每个表的状态
	tableStatus := make(map[string]string)
	var statusMutex sync.Mutex

	// 定期打印状态的goroutine
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				statusMutex.Lock()
				completed := 0
				failed := 0
				inProgress := 0
				for _, status := range tableStatus {
					switch status {
					case "completed":
						completed++
					case "failed":
						failed++
					case "in_progress":
						inProgress++
					}
				}
				statusMutex.Unlock()
				log.Printf("📊 进度统计: 完成 %d, 失败 %d, 进行中 %d, 总计 %d",
					completed, failed, inProgress, len(tables))
			case <-done:
				return
			}
		}
	}()

	for w := 0; w < *workerNum; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for tableName := range jobs {
				// 为每个表创建带超时的上下文
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
				migrateOneTableWithContext(ctx, id, dmConn, mysqlConn, tableName, tableStatus, &statusMutex)
				cancel()
			}
		}(w)
	}

	for _, t := range tables {
		jobs <- t
	}
	close(jobs)
	wg.Wait()
	close(done)

	mysqlConn.EnableConstraints()
	duration := time.Since(startTime)
	log.Printf("✅ 迁移完成，耗时: %v", duration)
}

func migrateOneTableWithContext(ctx context.Context, workerID int, dm *database.DMConnector, mysql *database.MySQLConnector, tableName string, tableStatus map[string]string, statusMutex *sync.Mutex) {
	// 使用select检查上下文是否已取消
	select {
	case <-ctx.Done():
		log.Printf("[Worker %d] ⚠️  表 %s 处理超时或被取消", workerID, tableName)
		statusMutex.Lock()
		tableStatus[tableName] = "failed"
		statusMutex.Unlock()
		return
	default:
	}

	statusMutex.Lock()
	tableStatus[tableName] = "in_progress"
	statusMutex.Unlock()

	startTime := time.Now()
	log.Printf("[Worker %d] 🔧 开始处理表 %s", workerID, tableName)

	// 在单独的goroutine中执行实际工作，并监听上下文取消信号
	done := make(chan error, 1)
	go func() {
		done <- migrateOneTableInternal(workerID, dm, mysql, tableName, tableStatus, statusMutex, startTime)
	}()

	select {
	case <-ctx.Done():
		log.Printf("[Worker %d] ⚠️  表 %s 处理超时", workerID, tableName)
		statusMutex.Lock()
		tableStatus[tableName] = "failed"
		statusMutex.Unlock()
	case err := <-done:
		if err != nil {
			log.Printf("[Worker %d] ❌ 表 %s 处理出错: %v", workerID, tableName, err)
			statusMutex.Lock()
			tableStatus[tableName] = "failed"
			statusMutex.Unlock()
		}
	}
}

func migrateOneTableInternal(workerID int, dm *database.DMConnector, mysql *database.MySQLConnector, tableName string, tableStatus map[string]string, statusMutex *sync.Mutex, startTime time.Time) error {
	// 逻辑同前，省略以节省篇幅...
	// 这里直接调用 mysql.CreateTable 和 mysql.BatchInsertData 即可
	dmCols, err := dm.GetTableSchema(tableName)
	if err != nil {
		log.Printf("[Worker %d] ❌ 获取结构失败 %s: %v", workerID, tableName, err)
		return err
	}

	log.Printf("[Worker %d] 📋 表 %s 包含 %d 个字段", workerID, tableName, len(dmCols))

	// 将 DMColumn 转换为 MySQLColumn
	mysqlCols := make([]database.MySQLColumn, len(dmCols))
	for i, col := range dmCols {
		mysqlCols[i] = database.MySQLColumn{
			Name:            col.Name,
			DataType:        col.DataType,
			DataLength:      col.DataLength,
			DataPrecision:   col.DataPrecision,
			DataScale:       col.DataScale,
			Nullable:        col.Nullable,
			IsPrimaryKey:    col.IsPrimaryKey,
			IsAutoIncrement: col.IsIdentity,
		}
	}

	log.Printf("[Worker %d] 🛠️  正在创建表 %s", workerID, tableName)
	if err := mysql.CreateTable(tableName, mysqlCols); err != nil {
		log.Printf("[Worker %d] ❌ 建表失败 %s: %v", workerID, tableName, err)
		return err
	}
	log.Printf("[Worker %d] ✅ 表 %s 创建成功", workerID, tableName)

	log.Printf("[Worker %d] 📥 正在读取表 %s 数据", workerID, tableName)
	rows, err := dm.GetTableData(tableName)
	if err != nil {
		log.Printf("[Worker %d] ❌ 读数据失败 %s: %v", workerID, tableName, err)
		return err
	}
	defer rows.Close()

	log.Printf("[Worker %d] 💾 正在写入表 %s 数据", workerID, tableName)
	insertedRows, err := mysql.BatchInsertData(tableName, mysqlCols, rows, *batchSize)
	if err != nil {
		log.Printf("[Worker %d] ❌ 写数据失败 %s: %v", workerID, tableName, err)
		return err
	}

	duration := time.Since(startTime)
	log.Printf("[Worker %d] ✅ %s 完成 (%d 行, 耗时: %v)", workerID, tableName, insertedRows, duration)

	statusMutex.Lock()
	tableStatus[tableName] = "completed"
	statusMutex.Unlock()

	return nil
}

func migrateOneTable(workerID int, dm *database.DMConnector, mysql *database.MySQLConnector, tableName string, tableStatus map[string]string, statusMutex *sync.Mutex) {
	// 保留此函数以保持向后兼容性，但实际逻辑已转移到带上下文的版本
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	migrateOneTableWithContext(ctx, workerID, dm, mysql, tableName, tableStatus, statusMutex)
	cancel()
}
