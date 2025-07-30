package main

import (
	"database/sql"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 连接MySQL数据库
	dsn := "root:870629@tcp(localhost:3306)/aimonitor"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 生成admin123的bcrypt哈希
	password := "admin123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to generate hash:", err)
	}

	fmt.Printf("Generated hash for '%s': %s\n", password, string(hash))
	fmt.Printf("Hash length: %d\n", len(string(hash)))

	// 更新数据库中的密码
	_, err = db.Exec("UPDATE users SET password = ? WHERE username = 'admin'", string(hash))
	if err != nil {
		log.Fatal("Failed to update password:", err)
	}

	fmt.Println("Password updated successfully!")

	// 验证更新
	var storedHash string
	err = db.QueryRow("SELECT password FROM users WHERE username = 'admin'").Scan(&storedHash)
	if err != nil {
		log.Fatal("Failed to query password:", err)
	}

	fmt.Printf("Stored hash: %s\n", storedHash)
	fmt.Printf("Stored hash length: %d\n", len(storedHash))

	// 测试密码验证
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err == nil {
		fmt.Printf("✓ Password verification successful!\n")
	} else {
		fmt.Printf("✗ Password verification failed: %v\n", err)
	}
}