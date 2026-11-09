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

type UserInput struct {
	Name     string    `joson:"name"`
	Email    string    `joson:"email"`
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
	fmt.Printf("user 2: %v\n", u)

	// 查询，使用泛型
	store := Store{ db: db }
	u, err = store.One(ctx,
		"SELECT id, name, email, created_at FROM users WHERE id = ?",
		scanUser,
		3,
	)

	if err != nil {
		log.Printf("failed to store.One, error: %v\n", err)
	}
	fmt.Printf("user 3: %v\n", u)
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
