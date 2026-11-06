package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID       int64     `joson:"id"`
	Name     string    `joson:"name"`
	Email    string    `joson:"email"`
	CreatedAt time.Time `joson:"created_at"`
}

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
	u, err := getUser(ctx, db, 2)
	if err != nil {
		log.Printf("failed to Scan(&name, &email), error: %v\n", err)
	}
	fmt.Printf("user: %v\n", u)
	// db.ExecContext()
}

func getUser(ctx context.Context, db *sql.DB, id int64) (User, error) {
	var u User
	err := db.QueryRowContext(
		ctx,
		"SELECT id, name, email, created_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	return u, err
}
