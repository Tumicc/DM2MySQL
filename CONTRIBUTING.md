# 贡献指南

感谢您对 DM2MySQL 项目的关注！

## 🤝 如何贡献

我们欢迎所有形式的贡献,包括但不限于:

- 🐛 报告 Bug
- 💡 提出新功能建议
- 📝 改进文档
- 🔧 提交代码修复
- 🌍 帮助翻译文档

## 🔐 安全第一 - 重要!

在提交任何内容之前,请**务必**确保:

### ❌ 绝对禁止包含的信息

1. **真实的生产环境数据**
   - 数据库密码、密钥、令牌
   - 内网 IP 地址或域名
   - 真实的表名、字段名
   - 包含个人信息的示例数据

2. **公司内部信息**
   - 内部网络拓扑
   - 服务器配置详情
   - 业务逻辑和流程

### ✅ 正确的示例数据格式

```json
// ❌ 错误示例 - 包含真实信息
{
  "tables": ["sys_user", "customer_info", "order_history"]
}

// ✅ 正确示例 - 使用通用表名
{
  "tables": ["users", "orders", "products", "employees"]
}
```

```bash
# ❌ 错误示例 - 真实连接信息
-dm-host=192.168.1.100
-dm-user=production_user
-dm-pass=RealPassword123

# ✅ 正确示例 - 占位符
-dm-host=your-dm-host
-dm-user=your-dm-user
-dm-pass=your-dm-password
```

### 🔍 提交前检查清单

在提交 PR 前,请确认:

- [ ] 所有敏感信息已被移除或替换为占位符
- [ ] 示例使用通用的表名(users, orders, products 等)
- [ ] IP 地址使用 `127.0.0.1` 或 `your-host`
- [ ] 密码使用 `your-password` 或环境变量示例
- [ ] 没有包含公司名称、项目名称的路径
- [ ] 日志文件不在提交范围内

## 📝 提交规范

### Commit Message 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型:**
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式(不影响功能)
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具相关

**示例:**
```
feat(batch): add automatic batch size calculation

- Calculate safe batch size based on column count
- Respect MySQL placeholder limit (65535)
- Add logging for batch optimization

Closes #123
```

## 🧪 测试指南

### 本地测试

在提交代码前,请确保:

1. **代码质量**
   ```bash
   # 格式化代码
   go fmt ./...

   # 静态检查
   go vet ./...

   # 运行测试(如果有)
   go test ./...
   ```

2. **功能测试**
   - 使用测试数据验证功能
   - 不要使用生产数据
   - 确保没有内存泄漏

3. **文档检查**
   - README 中的示例是否可运行
   - 参数说明是否准确
   - 没有拼写错误

## 📧 联系方式

如有疑问,请通过以下方式联系:

- **GitHub Issues**: [项目 Issues 页面]
- **Email**: noreply@example.com

## 📄 许可证

提交的贡献代码将采用与项目相同的 MIT 许可证。

---

**再次提醒: 请勿在公开平台发布任何敏感信息!** 🔒
