package main

import (
	"log"
	"net/http"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

/*
GORM 默认将结构体名转为蛇形命名的复数形式作为表名，因此：
- Usr → usrs
- User → users
*/
type Usr struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (u Usr) TableName() string {
	return "users"
}

func startServer() {
	log.Println("using gorm ...")

	dsn := "app:app_pass@tcp(127.0.0.1:3307)/appdb?parseTime=true&loc=UTC&timeout=5s&clientFoundRows=true" // 它让 MySQL 的 UPDATE 返回匹配行数，而不是实际发生变化的行数。否则对已有用户重复更新相同数据时，可能返回 0，导致程序错误地判断用户不存在。

	/*
		在完成初始化后，GORM 会自动 ping 数据库以检查数据库的可用性，若要禁用该特性，可将其设置为 true
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			DisableAutomaticPing: true,
		})
	*/
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// 学习阶段打印 SQL
	db = db.Debug()

	/*
		复用 usersDB 对 users 表进行增删改查
		这些调用的条件不会互相累积。 泛型 API 的 Where、`Order` 等方法会返回新的链式对象，原来的 usersDB 保持不变：
	*/
	usersDB := gorm.G[Usr](db)

	mux := http.NewServeMux()

	// 1. 创建 POST /users
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		in, ok := readInput(w, r)
		if !ok {
			return
		}

		user := Usr{Name: in.Name, Email: in.Email}
		/*
			&user 既提供输入，也接收回填结果:
			输入：GORM 读取 user.Name、user.Email 等字段，插入数据库。
			输出：GORM 将数据库生成的自增 ID 写回 user.ID，同时自动填充 CreatedAt
		*/
		err := usersDB.Create(r.Context(), &user)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, user)
	})

	// 2. 查询
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(w, r)
		if !ok {
			return
		}

		user, err := usersDB.Where("id = ?", id).First(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, user)
	})

	// 3. 更新
	mux.HandleFunc("PUT /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(w, r)
		if !ok {
			return
		}

		in, ok := readInput(w, r)
		if !ok {
			return
		}

		user := Usr{Name: in.Name, Email: in.Email}

		/*
			**1. `err == nil` 时，`rows` 可能为 `0` 吗？**

			可以。`err == nil` 表示执行没有出错，不代表找到了记录。例如 `WHERE id = ?` 没有匹配任何用户时，会返回 `rows == 0`，通常不会返回 `gorm.ErrRecordNotFound`。

			另外，如果结构体中没有任何需要更新的字段，GORM 也可能不执行 UPDATE，返回 `0`。

			**2. 是不是只更新 `user` 中的非零值字段？**

			是的，传入结构体时，默认只更新非零值字段。对这里的 `Usr`：

			- `Name != ""`：更新 `name`。
			- `Email != ""`：更新 `email`。
			- `ID == 0`、`CreatedAt` 为零时间：跳过。

			这里的“零值”指 **Go 类型的零值**，与数据库里的旧值无关。若希望空字符串也能写入，可显式选择字段：[GORM 更新文档](https://gorm.io/docs/update.html)。

			```go
			rows, err := usersDB.
					Where("id = ?", id).
					Select("Name", "Email").
					Updates(r.Context(), user)
			```

			**3. 字段和数据库原值一样，`rows` 返回 `0` 还是 `1`？**

			你当前的 DSN 配置了 **`clientFoundRows=true`**，因此返回的是匹配行数。假设该 ID 存在且确实执行了 UPDATE，即使值完全相同，也返回 **`1`**。

			| 情况 | 当前配置 `true` | 默认配置 `false` |
			|---|---:|---:|
			| ID 不存在 | 0 | 0 |
			| ID 存在，值发生变化 | 1 | 1 |
			| ID 存在，值完全相同 | 1 | 0 |

			这是 MySQL 驱动的配置行为：[go-sql-driver/mysql 文档](https://pkg.go.dev/github.com/go-sql-driver/mysql#readme-clientfoundrows)。
		*/
		rows, err := usersDB.Where("id = ?", id).Updates(r.Context(), user)
		if err != nil {
			writeError(w, err)
			return
		}
		if rows == 0 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("user not found"))
			return
		}
		w.Write([]byte("修改成功")) // 自动设置 status 200 
	})

	// 4. 查询列表
	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		list, err := usersDB.Order("id DESC").Limit(10).Find(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, list)
	})

	// 5. 删除
	mux.HandleFunc("DELETE /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(w, r)
		if !ok {
			return
		}

		rows, err := usersDB.Where("id = ?", id).Delete(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		if rows == 0 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("user not found"))
			return
		}
		w.Write([]byte("删除成功")) // 自动设置 status 200 
	})

	log.Println("HTTP server listening on 127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", mux))
}
