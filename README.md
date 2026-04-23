# 高并发秒杀选课系统 (High-Concurrency Course Selection)

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Redis](https://img.shields.io/badge/Redis-7.0+-DC382D?style=flat&logo=redis)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.12+-FF6600?style=flat&logo=rabbitmq)
![MySQL](https://img.shields.io/badge/MySQL-8.0+-4479A1?style=flat&logo=mysql)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

## 项目简介

本项目是一个基于 Golang 构建的高并发秒杀/选课系统，模拟高校海量学生同时抢课的瞬时并发场景。系统围绕选课核心链路，实现了鉴权、课程缓存、Redis Lua 原子预扣库存、Redis Stream 消息表、RabbitMQ 异步削峰、RabbitMQ 分级重试与死信队列、MySQL 事务落库与幂等控制。

**技术栈：** Golang, Gin, GORM, MySQL, Redis, RabbitMQ, Zap, Docker

## 核心架构与技术亮点

### 1. 接口鉴权与缓存防护

基于 Gin + JWT 实现无状态鉴权，核心选课接口通过布隆过滤器拦截不存在的课程 ID。课程详情查询使用“布隆过滤器 + 缓存空值 + singleflight”，降低缓存穿透和热点缓存击穿风险。

### 2. Redis Lua 原子预扣库存

选课请求进入核心链路后，通过 Redis Lua 在一次原子操作中完成重复请求判断、库存预扣减和 Redis Stream 消息写入。接口层只完成快速校验和入队，成功后返回“排队中”，把后续落库压力交给异步链路处理。

### 3. Redis Stream Outbox 保证消息可恢复

Redis Stream 作为轻量消息表，保存“已扣 Redis 库存但尚未确认投递 RabbitMQ”的消息副本。后台 relay 协程通过消费组读取 `select:stream`，投递 RabbitMQ 并等待 publisher confirm，收到 broker ack 后才执行 `XACK`；进程异常时，未确认消息会留在 pending list 中，后续通过 `XAUTOCLAIM` 回收重投。Stream 后台定期 `XTRIM`，控制消息表内存占用。

### 4. RabbitMQ 削峰、分级重试与死信队列

RabbitMQ 主队列负责异步削峰，消费者失败时不直接无限 requeue，而是按 `1s -> 5s -> 10s` 进入不同 TTL 重试队列，过期后自动路由回主队列再次消费。超过 3 次仍失败的消息进入死信队列，等待人工排查。消费者只有在重试消息或死信消息发布成功并收到 confirm 后，才 `Ack` 原消息，避免失败转发窗口丢消息。

### 5. MySQL 事务落库与一致性取舍

消费端使用 MySQL 事务扣减真实库存并创建选课记录，通过学生 ID + 课程 ID 唯一索引保证重复消息不会重复落库。Redis 开启 AOF `appendfsync everysec`，在性能和可靠性之间做折中；极端宕机场景理论上可能丢失约 1 秒内未刷盘的数据，如需更强一致性可改为 MySQL Outbox。

## 部署与性能压测

基于 Docker Compose 部署在 2C4G 云服务器，使用 `wrk` 对核心选课接口进行压测。压测前会清空 RabbitMQ 队列、刷新 Redis、重置 MySQL 课程库存，并重新启动应用以加载最新库存。

### 场景一：库存有限，快速拒绝超额请求

1 万独立 JWT Token 争抢 1000 库存，200 并发，持续 10 秒：

```bash
wrk -t4 -c200 -d10s --latency -s scripts/post.lua http://127.0.0.1:8080
```

| 指标 | 结果 |
| --- | --- |
| 总请求数 | 57174 |
| QPS | 5673.92 req/s |
| 平均延迟 | 35.47 ms |
| P50 | 34.58 ms |
| P75 | 43.93 ms |
| P90 | 53.19 ms |
| P99 | 80.32 ms |
| 非 2xx/3xx 响应 | 56174 |

该场景下成功请求数量与库存规模一致，其余请求在库存耗尽后被快速拒绝，主要用于验证 Redis Lua 预扣库存链路在高并发下的稳定性。

### 场景二：库存充足，验证入队吞吐

4 万独立 JWT Token，课程库存重置为 12000，100 并发，持续 2 秒：

```bash
go run scripts/gen_tokens.go 40000

docker compose stop app
docker compose exec rabbitmq rabbitmqctl purge_queue redisQueue
docker compose exec redis redis-cli -a redispassword FLUSHALL
docker compose exec mysql mysql -uroot -prootpassword -D go_course -e "
TRUNCATE TABLE selections;
UPDATE courses SET stock = 12000 WHERE id = 1;
"
docker compose start app
sleep 30

wrk -t4 -c100 -d2s --latency -s scripts/request.lua http://127.0.0.1:8080
```

| 指标 | 结果 |
| --- | --- |
| 总请求数 | 6817 |
| QPS | 3258.04 req/s |
| 平均延迟 | 32.06 ms |
| P50 | 29.08 ms |
| P75 | 41.75 ms |
| P90 | 53.39 ms |
| P99 | 84.91 ms |
| 非 2xx/3xx 响应 | 0 |

该场景用于验证库存充足时接口层完成 JWT 校验、Redis Lua 预扣库存和 Redis Stream 入队后的吞吐能力。RabbitMQ 与 MySQL 的最终落库由后台 relay 和消费者异步完成。

## 快速启动 (Quick Start)

### 1. 环境依赖

* Golang >= 1.21
* MySQL >= 8.0
* Redis >= 7.0
* RabbitMQ >= 3.12

### 2. 克隆与运行

```bash
# 1. 克隆项目
git clone https://github.com/1KURA-hub/course-select.git
cd course-select

# 2. 安装依赖
go mod tidy

# 3. 配置环境，请修改 config/config.yaml 中的中间件连接地址
# MySQL: root:123456@tcp(127.0.0.1:3306)/course_select
# Redis: 127.0.0.1:6379
# RabbitMQ: amqp://guest:guest@127.0.0.1:5672/

# 4. 启动服务
go run main.go
```

## CI/CD (GitHub Actions + SSH)

### 一次性配置

1. 在服务器确保可执行部署脚本：

```bash
cd /go-course/course-select
chmod +x scripts/deploy.sh
```

2. 在 GitHub 仓库配置 Actions Secrets：

- `SSH_HOST`：服务器公网 IP
- `SSH_USER`：服务器登录用户，如 `root`
- `SSH_PORT`：SSH 端口，通常为 `22`
- `SSH_KEY`：用于登录服务器的私钥内容
- `DEPLOY_PATH`：服务器项目路径，如 `/go-course/course-select`

3. 确保仓库 Actions 具有读写权限：

`Settings -> Actions -> General -> Workflow permissions -> Read and write permissions`

### 日常使用

每次推送到 `main` 会自动：

1. 执行 `go test ./...`
2. SSH 到服务器执行 `scripts/deploy.sh`
3. 完成 `git pull + docker compose up -d --build app + healthz 检查`

手动触发：`Actions -> CI-CD -> Run workflow`

## 许可证

本项目采用 [MIT License](https://opensource.org/licenses/MIT) 开源许可证。
