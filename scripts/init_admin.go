package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"exchange/internal/models/mysql"
	"exchange/internal/pkg/config"
	"exchange/internal/pkg/database"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化MySQL连接
	mysqlService, err := database.NewMySQLService(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer mysqlService.Close()

	// 获取数据库实例
	db := mysqlService.DB()

	// 创建admin用户
	admin := &mysql.Admin{
		Username: "admin",
		Email:    "admin@example.com",
		Role:     mysql.AdminRoleAdmin,
		Status:   mysql.AdminStatusActive,
	}

	// 设置密码
	if err := admin.SetPassword("admin123"); err != nil {
		log.Fatalf("Failed to set password: %v", err)
	}

	// 验证admin数据
	if err := admin.Validate(); err != nil {
		log.Fatalf("Admin validation failed: %v", err)
	}

	// 检查admin用户是否已存在
	var existingAdmin mysql.Admin
	result := db.Where("username = ?", admin.Username).First(&existingAdmin)
	if result.Error == nil {
		fmt.Println("Admin user already exists!")
		fmt.Printf("Username: %s\n", existingAdmin.Username)
		fmt.Printf("Email: %s\n", existingAdmin.Email)
		fmt.Printf("Role: %s\n", existingAdmin.Role)
		fmt.Printf("Status: %s\n", existingAdmin.Status)
		os.Exit(0)
	}

	// 创建admin用户
	ctx := context.Background()
	if err := db.WithContext(ctx).Create(admin).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Println("✅ Admin user created successfully!")
	fmt.Printf("Username: %s\n", admin.Username)
	fmt.Printf("Email: %s\n", admin.Email)
	fmt.Printf("Role: %s\n", admin.Role)
	fmt.Printf("Status: %s\n", admin.Status)
	fmt.Printf("Password: admin123\n")
	fmt.Println("\nYou can now login with:")
	fmt.Println("Username: admin")
	fmt.Println("Password: admin123")
}
