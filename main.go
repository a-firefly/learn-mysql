package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	println("hello, world!")
	
	db, err := sql.Open(
		"mysql",
		"app:app_pass@tcp(127.0.0.1:3307)/appdb?parseTime=true&loc=UTC",
	)
	if err != nil {
		log.Fatalf("failed to open db, error: %v\n", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("mysql 连接成功")

	ctx := context.Background()

	// 插入
	result, err := db.ExecContext(
		ctx,
		"INSERT INTO users(name, email) VALUES(?, ?)",
		"kobe",
		"kobe@qq.com",
	)

	if err != nil {
		log.Printf("failed to db.ExecContext, error: %v\n", err)
	} else {
		id, err := result.LastInsertId()
		if err != nil {
			fmt.Printf("result.LastInsertId(), error: %v\n", err)
		} else {
			fmt.Printf("LastInsertId: %d\n", id)
		}
	}

	// 查询
	var name string
	var email string
	var craetedAt time.Time
	err = db.QueryRowContext(
		ctx,
		"SELECT name, email, created_at FROM users WHERE id = ?",
		2,
	).Scan(&name, &email, &craetedAt)
	if err != nil {
		log.Printf("failed to Scan(&name, &email), error: %v\n", err)
	}
	fmt.Printf("user: id = 2, name: %s, email: %s, craeteAt: %v\n", name, email, craetedAt)
	// db.ExecContext()
}