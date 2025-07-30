package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ai-monitor/internal/models"
)

func main() {
	// 连接数据库
	db, err := gorm.Open(sqlite.Open("data/aimonitor.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 查找admin用户
	var user models.User
	err = db.Where("username = ?", "admin").First(&user).Error
	if err != nil {
		log.Fatal("Failed to find admin user:", err)
	}

	fmt.Printf("Admin User Info:\n")
	fmt.Printf("ID: %s\n", user.ID)
	fmt.Printf("Username: %s\n", user.Username)
	fmt.Printf("Email: %s\n", user.Email)
	fmt.Printf("Status: %s\n", user.Status)
	fmt.Printf("Password Hash: %s\n", user.Password)

	// 测试密码验证
	passwords := []string{"admin123", "admin", "password", "123456"}
	for _, pwd := range passwords {
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pwd))
		if err == nil {
			fmt.Printf("✓ Password '%s' matches!\n", pwd)
		} else {
			fmt.Printf("✗ Password '%s' does not match\n", pwd)
		}
	}

	// 生成新的admin123密码哈希用于比较
	newHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to generate hash:", err)
	}
	fmt.Printf("\nNew hash for 'admin123': %s\n", string(newHash))
}