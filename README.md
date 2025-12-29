<div align="center">
<h1>DM2MySQL - 达梦数据库到 MySQL 迁移工具</h1>

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![DM Version](https://img.shields.io/badge/DM-8.x-blue.svg)](https://www.dameng.com/)
[![MySQL Version](https://img.shields.io/badge/MySQL-5.x%20%7C%208.x-orange.svg)](https://www.mysql.com/)

**高性能 · 生产级 · 易用性**

一个专为国产化替代场景打造的企业级数据库迁移工具

[功能特性](#功能特性) • [快速开始](#快速开始) • [使用文档](#使用文档) • [性能调优](#性能调优) • [常见问题](#常见问题)

</div>

---

## 📋 目录

- [项目背景](#项目背景)
- [核心功能](#核心功能)
- [技术架构](#技术架构)
- [系统要求](#系统要求)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [使用文档](#使用文档)
- [性能调优](#性能调优)
- [数据类型映射](#数据类型映射)
- [常见问题](#常见问题)
- [最佳实践](#最佳实践)
- [贡献指南](#贡献指南)
- [许可证](#许可证)

---

## 🎯 项目背景

在国产化替代和信创产业快速发展的背景下,许多企业需要将原有的达梦数据库(DM)迁移到 MySQL 生态系统,以获得更广泛的社区支持、更好的云原生兼容性和更低的运维成本。

**DM2MySQL** 是在真实的企业迁移项目中诞生的工具,针对以下痛点提供了解决方案:

- ✅ **跨平台兼容**: 达梦数据库主要用于中国本土项目,迁移到 MySQL 可获得更好的国际化支持
- ✅ **云原生适配**: MySQL 在各大云平台上有完善的托管服务,而达梦数据库支持有限
- ✅ **人才生态**: MySQL 技术栈的人才储备更丰富,降低企业人力成本
- ✅ **工具链完善**: MySQL 拥有丰富的监控、备份、高可用工具
- ✅ **成本优化**: 开源替代方案可显著降低数据库授权成本

> **适用场景**: 
> - 国产化替代项目中的数据库迁移
> - 传统行业数字化转型
> - 从专有数据库迁移到开源生态
> - 数据中台建设中的数据统一

---

## ⭐ 核心功能

### 1️⃣ 全自动迁移流程

```mermaid
graph LR
A[达梦数据库] --> B[提取表结构]
B --> C[类型映射转换]
C --> D[创建MySQL表]
D --> E[流式读取数据]
E --> F[批量写入MySQL]
F --> G[验证与完成]
```

- 🔁 **结构 + 数据全迁移**: 一次运行完成表结构和数据的完整迁移
- 🎯 **选择性迁移**: 通过配置文件指定需要迁移的表
- 🛡️ **安全第一**: 自动检测并处理表冲突,支持 `DROP TABLE IF EXISTS`

### 2️⃣ 智能类型映射

完整支持达梦数据库(DM8)的所有数据类型到 MySQL 5.x/8.x 的智能映射:

| 达梦类型 | MySQL 类型 | 说明 |
|---------|-----------|------|
| `NUMBER(p,s)` | `TINYINT/SMALLINT/INT/BIGINT/DECIMAL` | 根据精度自动选择最优类型 |
| `VARCHAR2(n)` | `VARCHAR(n)` | 自动处理超长字符串转为 LONGTEXT |
| `CLOB` | `LONGTEXT` | 大文本对象 |
| `BLOB` | `LONGBLOB` | 二进制大对象 |
| `DATE` | `DATETIME` | 达梦 DATE 包含时间部分 |
| `TIMESTAMP` | `DATETIME/DATETIME(6)` | 8.0 版本支持微秒精度 |

### 3️⃣ 高性能并发处理

```go
// 并发迁移架构示意图
┌─────────────────────────────────────────────┐
│           Main Coordinator                  │
├─────────────────────────────────────────────┤
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐   │
│  │ W-1  │  │ W-2  │  │ W-3  │  │ W-4  │   │
│  │      │  │      │  │      │  │      │   │
│  └──────┘  └──────┘  └──────┘  └──────┘   │
│     ↓        ↓        ↓        ↓          │
│  ┌─────────────────────────────────────┐  │
│  │      Shared Connection Pool         │  │
│  └─────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

- ⚡ **多 Worker 并发**: 默认 4 个并发 Worker,可自定义调整
- 📦 **智能批量插入**: 根据列数自动计算最优批次大小,适配 MySQL 占位符限制
- 🔌 **连接池优化**: 生产级连接池配置,支持高并发连接复用
- ⏱️ **超时控制**: 每张表 30 分钟超时保护,防止长时间挂起

### 4️⃣ 企业级可靠性

- 🔄 **自动重试机制**: 连接断开时自动重试(最多 3 次)
- 🛡️ **错误隔离**: 单表失败不影响其他表的迁移
- 📊 **实时进度监控**: 每 30 秒输出一次迁移进度统计
- 🔒 **约束管理**: 自动禁用/启用外键和唯一性检查,提升导入速度

### 5️⃣ 多版本兼容

- 🎯 **MySQL 5.x 兼容**: 使用 `utf8` 字符集和 `DATETIME` 类型
- 🎯 **MySQL 8.x 优化**: 使用 `utf8mb4` 字符集、`DYNAMIC` 行格式和微秒精度
- 🔧 **自动版本检测**: 通过 `-mysql-ver` 参数指定目标版本

---

## 🏗️ 技术架构

### 技术栈

```
语言: Go 1.21+
驱动: 
  - 达梦: gitee.com/chunanyong/dm v1.8.21
  - MySQL: github.com/go-sql-driver/mysql v1.7.1
架构: 分层架构 (Config Layer → Database Layer → Application Layer)
```

### 项目结构

```
dm2mysql/
├── main.go                 # 主程序入口,命令行参数解析
├── config/                 # 配置管理
│   ├── config.go          # 配置加载逻辑
│   └── tables.json        # 表迁移配置文件
├── database/              # 数据库抽象层
│   ├── dm.go              # 达梦数据库连接器
│   └── mysql.go           # MySQL 数据库连接器
├── go.mod                 # Go 模块依赖
├── go.sum                 # 依赖版本锁定
└── README.md              # 项目文档
```

### 核心设计模式

1. **连接器模式**: `DMConnector` 和 `MySQLConnector` 封装数据库操作
2. **流式处理**: 大表数据流式读取,避免内存溢出
3. **批量处理**: 智能批次计算,平衡性能与稳定性
4. **上下文控制**: 使用 `context.Context` 实现超时和取消机制

---

## 📦 系统要求

### 运行环境

| 组件 | 要求 |
|------|------|
| **Go 语言** | Go 1.21 或更高版本 |
| **操作系统** | Linux / macOS / Windows |
| **内存** | 建议 4GB+ (取决于数据量) |
| **网络** | 能够同时访问达梦和 MySQL 数据库 |

### 数据库要求

| 数据库 | 版本 | 权限要求 |
|--------|------|----------|
| **达梦数据库(源)** | DM8 | SELECT 权限(读取表结构和数据) |
| **MySQL(目标)** | 5.5+ / 8.0+ | CREATE, INSERT, DROP 权限 |

### 网络配置

- 确保运行机器能够访问源达梦数据库(默认端口 5236)
- 确保运行机器能够访问目标 MySQL 数据库(默认端口 3306)
- 建议在局域网环境运行,以获得最佳迁移速度

---

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/Tumicc/DM2MySQL.git
cd dm2mysql
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置迁移表列表

编辑 `config/tables.json`,指定需要迁移的表:

```json
{
  "tables": [
    "cdm_codetype",
    "cdm_codedict",
    "knowledge",
    "rule"
    // ... 添加更多表名
  ]
}
```

### 4. 运行迁移

#### 基础用法(MySQL 5.x)

```bash
go run main.go \
  -dm-host=your-dm-host \
  -dm-port=5236 \
  -dm-user=your-dm-user \
  -dm-pass=your-dm-password \
  -dm-schema=your-dm-schema \
  -mysql-host=127.0.0.1 \
  -mysql-port=3306 \
  -mysql-user=root \
  -mysql-pass=your-mysql-password \
  -mysql-db=target_database \
  -mysql-ver=5
```

#### 高级用法(MySQL 8.x + 性能优化)

```bash
go run main.go \
  -dm-host=your-dm-host \
  -dm-port=5236 \
  -dm-user=your-dm-user \
  -dm-pass=your-dm-password \
  -dm-schema=your-dm-schema \
  -mysql-host=127.0.0.1 \
  -mysql-port=3306 \
  -mysql-user=root \
  -mysql-pass=your-mysql-password \
  -mysql-db=target_database \
  -mysql-ver=8 \
  -batch=5000 \
  -workers=8 \
  -tables-config=./config/tables.json
```

### 5. 验证迁移结果

连接到 MySQL 数据库,验证表结构和数据:

```sql
-- 查看已迁移的表
SHOW TABLES;

-- 检查数据行数
SELECT table_name, table_rows
FROM information_schema.tables
WHERE table_schema = 'your_database_name';

-- 验证表结构
DESC your_table_name;
```

---

## ⚠️ 安全提醒

### 🔐 敏感信息保护

在使用本工具进行数据库迁移时,请注意以下安全事项:

1. **不要在命令行中明文输入密码**
   ```bash
   # ❌ 不推荐: 密码会出现在命令历史和进程列表中
   -dm-pass=your-password

   # ✅ 推荐: 使用环境变量
   export DM_PASSWORD="your-password"
   -dm-pass=$DM_PASSWORD
   ```

2. **不要将包含真实数据的配置文件提交到版本控制系统**
   - `.gitignore` 已配置忽略敏感配置文件
   - 使用 `config/tables.json.example` 作为模板
   - 真实的 `config/tables.json` 应该在本地维护

3. **迁移前备份重要数据**
   ```bash
   # MySQL 备份
   mysqldump -u root -p --all-databases > backup_$(date +%Y%m%d).sql
   ```

4. **使用只读权限账户进行迁移**
   - 源数据库(达梦)只需要 SELECT 权限
   - 目标数据库(MySQL)需要 CREATE、INSERT、DROP 权限

5. **迁移完成后清理**
   - 删除或加密保存迁移日志
   - 修改数据库密码
   - 检查是否有临时文件残留

### 🚫 禁止公开的信息

以下信息**严禁**提交到 GitHub 或其他公开平台:

- ❌ 数据库连接字符串(包含 IP、端口、用户名、密码)
- ❌ 真实的业务表名和字段名
- ❌ 包含个人信息的示例数据
- ❌ 公司内部网络拓扑信息
- ❌ 任何形式的密钥或令牌

---

## ⚙️ 配置说明

### 命令行参数

#### 达梦数据库参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-dm-host` | string | `127.0.0.1` | 达梦数据库 IP 地址 |
| `-dm-port` | int | `5236` | 达梦数据库端口 |
| `-dm-user` | string | *必填* | 数据库用户名 |
| `-dm-pass` | string | *必填* | 数据库密码 |
| `-dm-schema` | string | *必填* | 模式名(类似于 MySQL 的 database) |
| `-dm-extra` | string | - | 额外连接参数(如 `timeout=30s`) |

#### MySQL 参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-mysql-host` | string | `127.0.0.1` | MySQL IP 地址 |
| `-mysql-port` | int | `3306` | MySQL 端口 |
| `-mysql-user` | string | `root` | MySQL 用户名 |
| `-mysql-pass` | string | *必填* | MySQL 密码 |
| `-mysql-db` | string | *必填* | 目标数据库名 |
| `-mysql-ver` | int | `5` | MySQL 版本: `5`(5.x) 或 `8`(8.0+) |
| `-mysql-extra` | string | - | 额外连接参数 |

#### 性能参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-batch` | int | `2000` | 批量插入的行数(建议 1000-10000) |
| `-workers` | int | `4` | 并发 Worker 数量(建议 4-16) |
| `-tables-config` | string | `./config/tables.json` | 表配置文件路径 |

### 表配置文件

`config/tables.json` 格式:

```json
{
  "tables": [
    "table1",
    "table2",
    "table3"
  ]
}
```

**提示**: 
- 如果要迁移所有表,可以使用 SQL 查询达梦数据库生成表列表
- 表名区分大小写,需要与达梦数据库中的实际表名一致

---

## 📚 使用文档

### 完整工作流程

```
1️⃣ 环境准备
   ├─ 安装 Go 1.21+
   ├─ 确保网络连通性
   └─ 备份目标数据库

2️⃣ 配置迁移任务
   ├─ 编辑 tables.json
   ├─ 准备连接参数
   └─ (可选)性能调优

3️⃣ 执行迁移
   ├─ 运行迁移命令
   ├─ 监控实时日志
   └─ 等待完成

4️⃣ 验证结果
   ├─ 检查表数量
   ├─ 验证数据行数
   ├─ 测试查询性能
   └─ 应用程序测试
```

### 日志输出说明

程序运行时会输出详细的日志信息:

```
2024/xx/xx xx:xx:xx 🚀 开始数据库迁移...
2024/xx/xx xx:xx:xx 🔗 正在连接到达梦数据库...
2024/xx/xx xx:xx:xx ✅ 达梦数据库连接成功
2024/xx/xx xx:xx:xx 🔗 正在连接到MySQL数据库...
2024/xx/xx xx:xx:xx 💡 检测到 MySQL 8.0+ 模式: 使用 utf8mb4 字符集
2024/xx/xx xx:xx:xx ✅ MySQL数据库连接成功
2024/xx/xx xx:xx:xx ⚙️  正在禁用约束检查...
2024/xx/xx xx:xx:xx ✅ 约束检查已禁用
2024/xx/xx xx:xx:xx 📋 准备迁移指定的 23 张表...
2024/xx/xx xx:xx:xx [Worker 1] 🔧 开始处理表 knowledge
2024/xx/xx xx:xx:xx [Worker 2] 🔧 开始处理表 rule
...
2024/xx/xx xx:xx:xx 📊 进度统计: 完成 15, 失败 0, 进行中 5, 总计 23
...
2024/xx/xx xx:xx:xx ✅ 迁移完成，耗时: 1h23m45s
```

### 错误处理

如果迁移过程中出现错误:

1. **查看详细日志**: 每个错误都会在日志中输出详细信息
2. **检查网络连接**: 确保能够访问两个数据库
3. **验证权限**: 确保数据库用户具有足够的权限
4. **单表重跑**: 由于错误隔离机制,可以修改 `tables.json` 只重跑失败的表

---

## ⚡ 性能调优

### 批次大小 (-batch)

| 场景 | 推荐值 | 说明 |
|------|--------|------|
| **普通表** | 2000-5000 | 平衡内存占用和性能 |
| **宽表**(列数 > 50) | 1000-2000 | 避免 MySQL 占位符限制 |
| **窄表**(列数 < 10) | 5000-10000 | 可以使用更大的批次 |
| **网络不稳定** | 500-1000 | 降低批次大小,减少重试影响 |

**技术细节**: 
- MySQL 预处理语句占位符限制为 65535
- 工具会自动计算: `safeBatchSize = 60000 / 列数`
- 实际使用的批次大小 = `min(用户指定, 安全计算值)`

### 并发数 (-workers)

| 机器配置 | 推荐值 | 说明 |
|---------|--------|------|
| **4核8G** | 4 | 默认配置,适合大多数场景 |
| **8核16G** | 8-12 | 高性能机器 |
| **16核32G+** | 16-24 | 企业级服务器 |
| **网络带宽有限** | 2-4 | 避免带宽竞争 |

**注意事项**:
- 过多的并发可能导致数据库连接池耗尽
- 建议根据数据库的 `max_connections` 参数调整
- 达梦数据库默认连接限制较严,需要提前确认

### 内存优化

对于超大表(>10GB),工具采用流式处理,内存占用恒定:

```
内存占用 ≈ (workers × batch_size × row_size) + connection_pool_overhead
```

例如: 4 workers, 5000 batch, 平均行大小 1KB
```
内存占用 ≈ 4 × 5000 × 1KB ≈ 20MB (可忽略不计)
```

### 网络优化

- 🏢 **局域网迁移**: 推荐在数据库服务器所在的局域网内运行
- 🌐 **跨公网迁移**: 建议使用 VPN 或专线,降低延迟
- 🔐 **防火墙配置**: 确保开放 5236(达梦) 和 3306(MySQL) 端口

---

## 🔄 数据类型映射

### 数值类型

| 达梦类型 | 精度/标度 | MySQL 5.x | MySQL 8.x | 说明 |
|---------|----------|-----------|-----------|------|
| `NUMBER(p,0)` | p ≤ 3 | `TINYINT` | `TINYINT` | -128 ~ 127 |
| `NUMBER(p,0)` | p ≤ 5 | `SMALLINT` | `SMALLINT` | -32768 ~ 32767 |
| `NUMBER(p,0)` | p ≤ 9 | `INT` | `INT` | -21亿 ~ 21亿 |
| `NUMBER(p,0)` | p ≤ 19 | `BIGINT` | `BIGINT` | int64 范围 |
| `NUMBER(p,0)` | p > 19 | `DECIMAL(p,0)` | `DECIMAL(p,0)` | 超大整数 |
| `NUMBER(p,s)` | s > 0 | `DECIMAL(p,s)` | `DECIMAL(p,s)` | 带小数 |
| `INTEGER` | - | `INT` | `INT` | 标准整型 |
| `BIGINT` | - | `BIGINT` | `BIGINT` | 长整型 |
| `SMALLINT` | - | `SMALLINT` | `SMALLINT` | 短整型 |
| `TINYINT` | - | `TINYINT` | `TINYINT` | 微整型 |
| `FLOAT/DOUBLE` | - | `DOUBLE` | `DOUBLE` | 双精度浮点 |

### 字符串类型

| 达梦类型 | 最大长度 | MySQL 类型 | 说明 |
|---------|----------|-----------|------|
| `VARCHAR2(n)` | n ≤ 5461 | `VARCHAR(n)` | 普通字符串 |
| `VARCHAR2(n)` | 5461 < n ≤ 21845 | `TEXT` | 中等长度文本 |
| `VARCHAR2(n)` | n > 21845 | `LONGTEXT` | 超长文本 |
| `CHAR(n)` | - | `CHAR(n)` | 固定长度 |
| `CLOB` | - | `LONGTEXT` | 大文本对象 |
| `TEXT` | - | `LONGTEXT` | 文本类型 |

**计算逻辑**:
- MySQL 单行最大约 65535 字节
- `utf8` 下: 1 字符 = 3 字节 → `VARCHAR(21845)` 为临界值
- `utf8mb4` 下: 1 字符 = 4 字节 → `VARCHAR(16383)` 为临界值

### 时间日期类型

| 达梦类型 | MySQL 5.x | MySQL 8.x | 说明 |
|---------|-----------|-----------|------|
| `DATE` | `DATETIME` | `DATETIME` | 达梦 DATE 包含时间部分 |
| `TIMESTAMP` | `DATETIME` | `DATETIME(6)` | 8.0 支持微秒精度 |
| `TIME` | `TIME` | `TIME` | 时间类型 |
| `DATETIME` | `DATETIME` | `DATETIME` | 日期时间 |

### 二进制类型

| 达梦类型 | MySQL 类型 | 说明 |
|---------|-----------|------|
| `BLOB` | `LONGBLOB` | 二进制大对象 |
| `IMAGE` | `LONGBLOB` | 图像数据 |
| `BINARY` | `LONGBLOB` | 固定长度二进制 |

### 特殊类型

| 达梦类型 | MySQL 类型 | 说明 |
|---------|-----------|------|
| `BIT` | `TINYINT(1)` | 布尔值(业界通用做法) |
| `BOOL/BOOLEAN` | `TINYINT(1)` | 布尔值 |

---

## ❓ 常见问题

### Q1: 迁移时提示 "连接被拒绝"

**原因**: 
- 网络不通
- 防火墙阻止
- 数据库服务未启动

**解决方案**:
```bash
# 测试网络连通性
telnet <dm-host> 5236
telnet <mysql-host> 3306

# 检查防火墙
sudo ufw status  # Linux
sudo firewall-cmd --list-all  # CentOS

# 检查数据库服务状态
# 达梦数据库
ps -ef | grep dm_service

# MySQL
systemctl status mysql
```

### Q2: 迁移后数据行数不一致

**可能原因**:
- 某些行因为格式错误被跳过
- 字符集问题导致数据截断
- 触发了目标数据库的约束检查

**解决方案**:
```sql
-- 1. 检查错误日志,查找失败的表
grep "❌" migration.log

-- 2. 对比行数
-- 达梦数据库
SELECT COUNT(*) FROM table_name;

-- MySQL
SELECT COUNT(*) FROM table_name;

-- 3. 检查是否有数据被截断
SELECT MAX(LENGTH(column_name)) FROM table_name;
```

### Q3: 字符串乱码问题

**原因**: 字符集不匹配

**解决方案**:
```bash
# 1. 确保 MySQL 连接字符串指定了正确的字符集
# MySQL 5.x
-mysql-ver=5  # 自动使用 utf8

# MySQL 8.x
-mysql-ver=8  # 自动使用 utf8mb4

# 2. 手动指定字符集
-mysql-extra="charset=utf8mb4"
```

### Q4: 批量插入失败 "incorrect parameters count"

**原因**: 批次大小超出 MySQL 占位符限制

**解决方案**:
- 工具会自动调整批次大小
- 如果仍然失败,手动调小 `-batch` 参数:
```bash
-batch=1000  # 从 2000 降低到 1000
```

### Q5: 迁移速度很慢

**优化建议**:
```bash
# 1. 增加并发数
-workers=8  # 从默认 4 增加到 8

# 2. 增加批次大小
-batch=5000  # 从默认 2000 增加到 5000

# 3. 确保在局域网环境运行
# 避免跨公网迁移

# 4. 检查网络带宽
# Linux
iftop -i eth0

# macOS
nettop -w
```

### Q6: 内存占用过高

**原因**: 批次大小设置过大

**解决方案**:
```bash
# 降低批次大小
-batch=500  # 从默认 2000 降低到 500

# 降低并发数
-workers=2  # 从默认 4 降低到 2
```

### Q7: 如何只迁移表结构,不迁移数据?

**当前版本**: 工具默认同时迁移结构和数据

**解决方案**: 
修改 `main.go` 中的 `migrateOneTableInternal` 函数,注释掉数据迁移部分:

```go
// 注释掉以下代码即可
// rows, err := dm.GetTableData(tableName)
// ...
// insertedRows, err := mysql.BatchInsertData(...)
```

或者在迁移完成后,立即清空数据:

```sql
-- MySQL 数据库执行
TRUNCATE TABLE table_name;
```

### Q8: 迁移大表(>100GB)时超时

**原因**: 默认单表超时时间为 30 分钟

**解决方案**:
修改 `main.go` 中的超时时间:

```go
// 从 30 分钟改为 120 分钟
ctx, cancel := context.WithTimeout(context.Background(), 120*time.Minute)
```

---

## 🎯 最佳实践

### 迁移前准备

1. **📋 数据评估**
   ```sql
   -- 达梦数据库执行
   SELECT 
     table_name,
     num_rows * avg_row_len / 1024 / 1024 AS size_mb
   FROM user_tables
   ORDER BY size_mb DESC;
   ```

2. **💾 备份目标数据库**
   ```bash
   # MySQL 备份
   mysqldump -u root -p --all-databases > backup.sql
   ```

3. **🧪 小规模测试**
   ```json
   // config/tables.json
   {
     "tables": ["small_table1", "small_table2"]
   }
   ```

### 迁移中监控

1. **📊 实时监控**
   ```bash
   # 监控 MySQL 连接数
   mysql -e "SHOW PROCESSLIST;" | wc -l

   # 监控磁盘 I/O
   iostat -x 1

   # 监控网络流量
   iftop -i eth0
   ```

2. **📝 日志记录**
   ```bash
   # 将日志输出到文件
   go run main.go [...] 2>&1 | tee migration.log
   ```

### 迁移后验证

1. **✅ 结构验证**
   ```sql
   -- 检查所有表是否创建成功
   SELECT COUNT(*) FROM information_schema.tables 
   WHERE table_schema = 'your_database';
   ```

2. **✅ 数据量验证**
   ```sql
   -- 检查每张表的行数
   SELECT table_name, table_rows 
   FROM information_schema.tables 
   WHERE table_schema = 'your_database'
   ORDER BY table_rows DESC;
   ```

3. **✅ 业务验证**
   - 运行应用程序的测试用例
   - 执行关键业务查询
   - 验证报表数据一致性

---

## 🤝 贡献指南

我们欢迎所有形式的贡献!

### 如何贡献

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 开发规范

- **代码风格**: 遵循 [Effective Go](https://golang.org/doc/effective_go.html) 指南
- **提交信息**: 使用清晰的提交信息格式
- **测试**: 添加新功能时请附带测试用例
- **文档**: 更新相关文档

### 待办事项

- [ ] 支持增量迁移(基于时间戳)
- [ ] 支持断点续传
- [ ] 支持数据校验和(MD5/SHA256)
- [ ] 提供 Web UI 界面
- [ ] 支持更多数据库类型(PostgreSQL, Oracle)

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

## 📮 联系方式

- **问题反馈**: [GitHub Issues](https://github.com/Tumicc/DM2MySQL/issues)
- **功能建议**: [GitHub Discussions](https://github.com/Tumicc/DM2MySQL/discussions)
- **邮件**: tumicc996@outlook.com

---

## 🙏 致谢

- [达梦数据库](https://www.dameng.com/) - 国产数据库优秀代表
- [Go 语言社区](https://golang.org/) - 提供强大的开发工具
- [MySQL](https://www.mysql.com/) - 全球最受欢迎的开源数据库

---

## 📊 项目统计

<div align="center">

![GitHub stars](https://img.shields.io/github/stars/Tumicc/DM2MySQL?style=social)
![GitHub forks](https://img.shields.io/github/forks/Tumicc/DM2MySQL?style=social)
![GitHub issues](https://img.shields.io/github/issues/Tumicc/DM2MySQL)
![GitHub license](https://img.shields.io/github/license/Tumicc/DM2MySQL)

**如果这个项目对您有帮助,请给一个 ⭐️ Star 支持一下!**

</div>
