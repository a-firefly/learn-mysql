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