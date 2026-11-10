package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID        int64     `joson:"id"`
	Name      string    `joson:"name"`
	Email     string    `joson:"email"`
	CreatedAt time.Time `joson:"created_at"`
}

type UserInput struct {
	Name  string `joson:"name"`
	Email string `joson:"email"`
}

type TransferInput struct {
	FromID int64 `json:"from_id"`
	ToID   int64 `json:"to_id"`
	Cents  int64 `json:"cents"`
}

func main() {
	println("hello, world!")

	db, err := sql.Open(
		"mysql",
		"app:app_pass@tcp(127.0.0.1:3307)/appdb?parseTime=true&loc=UTC&timeout=5s&clientFoundRows=true", // 它让 MySQL 的 UPDATE 返回匹配行数，而不是实际发生变化的行数。否则对已有用户重复更新相同数据时，可能返回 0，导致程序错误地判断用户不存在。
	)
	if err != nil {
		log.Fatalf("failed to open db, error: %v\n", err)
	}
	defer db.Close()

	// 连接池
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(3 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 啥意思

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("mysql 连接成功")

	// 查询，使用泛型
	store := Store{db: db}

	mux := http.NewServeMux() // 多路复用器

	const userColumns = "id, name, email, created_at"

	// 获取用户列表
	// GET /users
	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		list, err := store.All(
			r.Context(),
			"SELECT "+userColumns+" FROM users ORDER BY id DESC LIMIT 100",
			scanUser,
		)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, list)
	})

	// 获取指定用户
	// GET /users/{id}
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(w, r)
		if !ok {
			return
		}

		user, err := store.One(
			r.Context(),
			"SELECT "+userColumns+" FROM users WHERE id = ?",
			scanUser,
			id,
		)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, user)
	})

	// POST /usres
	// 注册用户
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		in, ok := readInput(w, r)
		if !ok {
			return
		}

		result, err := db.ExecContext(
			r.Context(),
			"INSERT INTO users(name, email) VALUES (?, ?)",
			in.Name, in.Email,
		)
		if err != nil {
			writeError(w, err)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			writeError(w, err)
			return
		}

		user, err := store.One(
			r.Context(),
			"SELECT "+userColumns+" FROM users WHERE id = ?",
			scanUser,
			id,
		)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, user)
	})

	// 修改指定用户信息
	// PUT /users/{id}
	mux.HandleFunc("PUT /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(w, r)
		if !ok {
			return
		}

		in, ok := readInput(w, r)
		if !ok {
			return
		}

		result, err := db.ExecContext(
			r.Context(),
			"UPDATE users SET name = ?, email = ? WHERE id = ?",
			in.Name, in.Email, id,
		)
		if err != nil {
			writeError(w, err)
			return
		}

		n, err := result.RowsAffected()
		if err != nil {
			writeError(w, err)
			return
		}
		if n == 0 {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		user, err := store.One(
			r.Context(),
			"SELECT "+userColumns+" FROM users WHERE id = ?",
			scanUser,
			id,
		)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, user)
	})

	// 删除指定用户
	// DELETE /users/{id}
	mux.HandleFunc("DELETE /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(w, r)
		if !ok {
			return
		}

		result, err := db.ExecContext(
			r.Context(),
			"DELETE FROM users WHERE id = ?",
			id,
		)
		if err != nil {
			writeError(w, err)
			return
		}

		n, err := result.RowsAffected()
		if err != nil {
			writeError(w, err)
			return
		}
		if n == 0 {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// 转账
	// POST /transfer
	mux.HandleFunc("POST /transfer", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()

		var in TransferInput
		err := decoder.Decode(&in)
		if err != nil {
			writeError(w, err)
			return
		}

		err = transfer(r.Context(), db, in.FromID, in.ToID, in.Cents)
		if err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("转账成功"))
	})

	log.Println("HTTP server listening on 127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", mux))
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

func transfer(
	ctx context.Context,
	db *sql.DB,
	fromID, toID int64,
	cents int64,
) error {
	if cents <= 0 || fromID == toID {
		return errors.New("invalid transfer")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 扣款
	result, err := tx.ExecContext(
		ctx,
		`UPDATE accounts
		SET balance_cents = balance_cents - ?
		WHERE id = ? AND balance_cents >= ?`, // 防止余额不足
		cents,
		fromID,
		cents,
	)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("insufficient balance or account missing")
	}

	// 收款
	result, err = tx.ExecContext(
		ctx,
		`UPDATE accounts
		SET balance_cents = balance_cents + ?
		WHERE id = ?`,
		cents,
		toID,
	)
	if err != nil {
		return err
	}

	n, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("target account missing")
	}

	return tx.Commit()
}
