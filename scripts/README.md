# 数据库初始化脚本

这个目录包含了用于初始化数据库和创建默认用户的脚本。

## 文件说明

### `init_admin.go`
- **功能**: 创建默认的admin用户
- **用户名**: admin
- **密码**: admin123
- **邮箱**: admin@example.com
- **角色**: admin
- **状态**: active

### `create_admin.sh`
- **功能**: 快速创建admin用户的shell脚本
- **用法**: `./scripts/create_admin.sh`

## 使用方法

### 方法1: 直接运行Go脚本
```bash
# 编译并运行
go build -o init_admin scripts/init_admin.go
./init_admin
```

### 方法2: 使用shell脚本
```bash
# 运行shell脚本（推荐）
./scripts/create_admin.sh
```

## 输出示例

成功创建admin用户后，您会看到类似以下的输出：

```
✅ Admin user created successfully!
Username: admin
Email: admin@example.com
Role: admin
Status: active
Password: admin123

You can now login with:
Username: admin
Password: admin123
```

## 登录信息

创建完成后，您可以使用以下信息登录admin面板：

- **登录地址**: `http://localhost:8080/admin/v1/auth/login`
- **用户名**: `admin`
- **密码**: `admin123`

## 注意事项

1. 脚本会自动检查admin用户是否已存在，如果已存在则不会重复创建
2. 密码必须符合安全要求（至少6个字符，包含字母和数字）
3. 确保数据库连接正常后再运行脚本
4. 脚本会自动处理密码加密和验证

## 故障排除

如果遇到问题，请检查：

1. 数据库连接是否正常
2. 配置文件是否正确
3. 数据库表是否已创建
4. 网络连接是否正常 