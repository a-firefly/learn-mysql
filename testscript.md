先启动 MySQL 和 HTTP 服务（默认地址 `http://127.0.0.1:8080`），然后运行：

```sh
make list
make 1
make create lily lily@qq.com
make update 1 lily lily@qq.com
make delete 1
```

直接映射到对应的 curl 命令，运行 `make` 默认查询用户列表。

可覆盖服务地址：

```sh
make list BASE_URL=http://127.0.0.1:8080
```

每次运行一个操作，按示例提供完整参数。姓名和邮箱使用不含空格、引号等特殊字符的简单值。


### 创建用户
curl -X POST http://127.0.0.1:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'

### 查询用户列表
curl http://127.0.0.1:8080/users

### 根据 ID 查询
curl http://127.0.0.1:8080/users/1

### 更新用户
curl -X PUT http://127.0.0.1:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Chen","email":"alice@example.com"}'

### 删除用户
curl -X DELETE http://127.0.0.1:8080/users/1