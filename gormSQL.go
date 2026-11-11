package main

import (
	"context"
	"fmt"
	"log"
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
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:100;not null"`
	Email     string    `gorm:"size:191;not null;uniqueIndex:uk_users_emial"`
	CreatedAt time.Time
}

func (u Usr) TableName() string {
	return "users"
}

func startServer() {
	fmt.Println("using gorm ...")
	log.Println("using gorm ...")

	dsn := "app:app_pass@tcp(127.0.0.1:3307)/appdb?parseTime=true&loc=UTC&timeout=5s&clientFoundRows=true" // 它让 MySQL 的 UPDATE 返回匹配行数，而不是实际发生变化的行数。否则对已有用户重复更新相同数据时，可能返回 0，导致程序错误地判断用户不存在。

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// 学习阶段打印 SQL
	db = db.Debug()

	ctx := context.Background()

	// 1. 创建用户
	user := Usr{
		Name: "alice",
		Email: fmt.Sprintf("alice-%dqq.com", time.Now().UnixNano()),
	}

	// 复用 usersDB 完成 users 表的增删改查
	// 这些调用的条件不会互相累积。 泛型 API 的 Where、`Order` 等方法会返回新的链式对象，原来的 usersDB 保持不变：
	usersDB := gorm.G[Usr](db)

	/*
	&user 既提供输入，也接收回填结果:
	输入：GORM 读取 user.Name、user.Email 等字段，插入数据库。
	输出：GORM 将数据库生成的自增 ID 写回 user.ID，同时自动填充 CreatedAt
	*/
	err = usersDB.Create(ctx, &user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created:", user)

	// 2. 查询
	found, err := usersDB.Where("id = ?", user.ID).First(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("found:", found.Name)

	// 3. 更新
	rows, err := usersDB.Where("id = ?", user.ID).Update(ctx, "name", "alice chen")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("updated rows:", rows)

	// 4. 查询列表
	users, err := usersDB.
			Order("id DESC").
			Limit(10).
			Find(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Users:", len(users))

	// 5. 删除
	rows, err = usersDB.Where("id = ?", user.ID).Delete(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted rows:", rows)

}
