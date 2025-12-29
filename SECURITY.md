# 安全政策

## 🔒 安全承诺

DM2MySQL 项目致力于保护用户的数据安全和隐私。我们非常重视安全问题,并承诺及时处理安全漏洞。

## 🚨 安全漏洞报告

### 如何报告

如果您发现安全问题,**请不要**公开提交 Issue。

**请通过以下方式私密报告:**

- 📧 Email: your-email@example.com
- 📢 使用 GitHub 的 ["Private vulnerability reporting"](https://github.com/your-org/dm2mysql/security/advisories) 功能

### 报告内容应包含

请尽可能提供以下信息:

- 漏洞描述和影响范围
- 复现步骤
- 受影响的版本
- 建议的修复方案(如果有)
- 您的联系信息

### 响应承诺

- 我们会在 **48 小时内** 确认收到报告
- 在 **7 天内** 提供初步评估
- 在合理的期限内发布修复补丁
- 修复发布前会通知报告者

## ✅ 安全最佳实践

### 对于用户

#### 1. 数据库连接安全

```bash
# ❌ 不安全: 密码明文出现在命令历史
go run main.go -dm-pass=MyPassword123

# ✅ 安全: 使用环境变量
export DM_PASSWORD="MyPassword123"
go run main.go -dm-pass=$DM_PASSWORD

# ✅ 更安全: 使用密钥管理工具
# AWS Secrets Manager / HashiCorp Vault 等
```

#### 2. 权限控制

**源数据库(达梦)最小权限:**
```sql
-- 只授予只读权限
GRANT SELECT ON schema_name TO username;
```

**目标数据库(MySQL)最小权限:**
```sql
-- 只授予必要的权限
GRANT CREATE, INSERT, DROP, SELECT, DELETE
ON database_name.* TO 'username'@'host';
```

#### 3. 网络安全

- 🏢 **推荐**: 在内网环境运行迁移
- 🔐 **加密**: 跨公网迁移必须使用 VPN 或 SSL/TLS
- 🚫 **隔离**: 避免在生产数据库高峰期运行

#### 4. 数据备份

```bash
# 迁移前必须备份!
# MySQL 备份
mysqldump -u root -p --all-databases \
  --single-transaction \
  --quick \
  --lock-tables=false \
  > backup_$(date +%Y%m%d_%H%M%S).sql

# 验证备份文件
ls -lh backup_*.sql
```

#### 5. 日志管理

```bash
# 迁移日志可能包含敏感信息
go run main.go [...] 2>&1 | tee migration.log

# 迁移完成后安全处理
# 加密保存
openssl enc -aes-256-cbc -salt -in migration.log -out migration.log.enc

# 或安全删除(确认不再需要)
shred -u migration.log
```

### 对于开发者

#### 1. 代码安全

- ❌ 不要在代码中硬编码凭据
- ❌ 不要将密码写入日志
- ✅ 使用环境变量或配置文件(并在 .gitignore 中)
- ✅ 敏感信息使用脱敏处理

#### 2. 依赖管理

```bash
# 定期检查依赖漏洞
go mod verify
go get -u all
go mod tidy

# 使用安全扫描工具
# govital (https://github.com/owenrumney/go-squirrel)
```

#### 3. SQL 注入防护

本项目使用参数化查询,天然防护 SQL 注入:

```go
// ✅ 安全: 使用参数化查询
rows, err := db.Query("SELECT * FROM ?", tableName)

// ❌ 危险: 拼接 SQL(不要这样做!)
query := fmt.Sprintf("SELECT * FROM %s", tableName)
```

## 🔐 敏感信息检查清单

在提交代码、Issue 或 PR 前,请检查:

- [ ] 无数据库密码、密钥、令牌
- [ ] 无真实 IP 地址、域名
- [ ] 无真实表名、字段名
- [ ] 无个人身份信息(PII)
- [ ] 无公司内部信息
- [ ] 日志文件已脱敏
- [ ] 配置文件已使用占位符

## 🛡️ 已知安全限制

### 当前版本的限制

1. **连接字符串**: 密码通过命令行传递,可能出现在进程列表中
   - **缓解措施**: 使用环境变量

2. **数据传输**: 默认不加密传输
   - **缓解措施**: 在内网运行或使用 VPN

3. **日志文件**: 可能包含表名和数据行数
   - **缓解措施**: 迁移后加密或删除日志

### 未来改进

- [ ] 支持配置文件加密
- [ ] 支持 SSL/TLS 数据传输
- [ ] 支持密钥管理服务集成
- [ ] 添加数据脱敏选项

## 📋 安全更新

### 建议订阅

如果您担心安全问题,建议:

1. **Watch GitHub Repository** - 接收安全更新通知
2. **订阅 Release 通知** - 及时获取安全补丁
3. **定期检查 CHANGELOG.md** - 了解安全修复

### 安全更新流程

```
发现漏洞 → 私密报告 → 确认评估 → 修复开发 → 测试验证 → 发布更新
```

## 📞 紧急联系

如需报告紧急安全问题:

- 📧 Email: your-email@example.com
- 🔑 PGP Key: [您的 PGP 公钥]

---

**最后更新**: 2024-XX-XX

**安全是每个人的责任!** 🔒
