#!/bin/bash

# 创建admin用户的快速脚本
echo "🚀 Creating admin user..."

# 编译并运行初始化脚本
go build -o init_admin scripts/init_admin.go
if [ $? -eq 0 ]; then
    echo "✅ Script compiled successfully"
    ./init_admin
    echo ""
    echo "🎉 Admin user initialization completed!"
    echo ""
    echo "📋 Login Information:"
    echo "   Username: admin"
    echo "   Password: admin123"
    echo "   Email: admin@example.com"
    echo ""
    echo "🔗 You can now login to the admin panel at:"
    echo "   http://localhost:8080/admin/v1/auth/login"
    echo ""
else
    echo "❌ Failed to compile script"
    exit 1
fi 