先启动 MySQL 和 HTTP 服务（默认地址 `http://127.0.0.1:8080`），然后运行：

```sh
make list
make 1
make create lily lily@qq.com
make update 1 lily lily@qq.com
make delete 1
make transfer 1 2 100
```

直接映射到对应的 curl 命令，运行 `make` 默认查询用户列表。

可覆盖服务地址：

```sh
make list BASE_URL=http://127.0.0.1:8080
```

每次运行一个操作，按示例提供完整参数。姓名和邮箱使用不含空格、引号等特殊字符的简单值。

`make transfer 1 2 100` 从账户 1 向账户 2 转账 100 分；三个参数均为整数，JSON 中使用数字，不加双引号。


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

### 这里的 `-X` 是 `curl` 的参数，用来**指定 HTTP 请求方法**。

文档里的例子：

- `-X POST`：发送 POST 请求，创建用户。
- `-X PUT`：发送 PUT 请求，更新用户。
- `-X DELETE`：发送 DELETE 请求，删除用户。

例如：

```bash
curl -X DELETE http://127.0.0.1:8080/users/1
```

表示向这个地址发送 DELETE 请求，删除 ID 为 `1` 的用户。

普通 `curl URL` 默认使用 GET；使用 `-d` 发送数据时，默认会使用 POST，所以创建用户示例里的 `-X POST` 可以省略。