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

### 1. 安全鉴权与入口拦截

基于 Gin 构建 HTTP API，使用 JWT 实现无状态鉴权。核心选课接口通过布隆过滤器快速拦截不存在的课程 ID，减少无效请求对缓存和数据库的冲击。

### 2. 缓存防护

课程详情查询采用“布隆过滤器 + 缓存空值”降低缓存穿透风险，并结合 `singleflight` 合并热点课程的并发回源请求，缓解缓存击穿时的数据库压力。

### 3. Redis Lua 原子预扣库存与消息表

选课请求进入核心链路后，使用 Redis Lua 脚本在一次原子操作中完成重复请求判断、库存预扣减和 Redis Stream 消息写入。请求成功后先返回“排队中”，后台 relay 协程再从 Redis Stream 读取消息并投递 RabbitMQ。

### 4. Redis Stream Outbox 与 RabbitMQ 发布确认

系统使用 Redis Stream 作为轻量 Outbox，解决“Redis 库存已扣减但应用进程未发出 MQ 消息就崩溃”的断点问题。relay 协程通过消费组读取 `select:stream`，投递 RabbitMQ 后等待 publisher confirm，只有收到 broker ack 后才执行 `XACK`。如果 relay 进程崩溃，未确认消息会留在 Stream pending list 中，后续通过 `XAUTOCLAIM` 回收并重试投递。

### 5. Redis Stream 裁剪与内存控制

`XACK` 只会移除消费组里的 pending 记录，不会删除 Stream 消息本体。为了避免 `select:stream` 无限增长，relay 会启动后台裁剪协程，每分钟执行一次 `XTRIM`。没有 pending 消息时，按 `MAXLEN ~ 500000` 保留最近约 50 万条消息；存在 pending 消息时，先查询最老 pending 消息 ID，再按 `MINID ~ oldestPendingID` 只裁剪它之前的历史消息，避免删除仍可能通过 `XAUTOCLAIM` 重试的消息。

生产环境中，Stream 保留长度应根据峰值入队 QPS、最大可接受下游积压时间和安全系数设置。例如峰值 3000 QPS，如果希望 RabbitMQ 或消费者故障 10 分钟内不丢本地消息，安全系数取 2，则需要保留约 `3000 * 600 * 2 = 3600000` 条消息。

### 6. RabbitMQ 分级重试与死信队列

RabbitMQ 使用一个 `direct` 交换机承载选课消息路由。正常消息通过 `select.main` 路由键进入主队列，只有主队列会被业务消费者消费。消费者落库失败时，不再直接无限 `Reject(true)`，而是把失败消息重新发布到同一个交换机，并根据 `x-retry-count` 选择不同重试路由键：第一次失败进入 1 秒重试队列，第二次失败进入 5 秒重试队列，第三次失败进入 10 秒重试队列。

三个重试队列只负责延迟，不被业务消费者直接消费。它们通过 `x-message-ttl` 设置不同过期时间，并通过 `x-dead-letter-exchange` 和 `x-dead-letter-routing-key` 在消息过期后自动把消息路由回主队列。消费者重新消费时会继续读取消息头里的重试次数；如果超过 3 次仍失败，消息会通过 `select.dlq` 路由到死信队列，等待人工排查和补偿。

消费者转发失败消息时会等待 RabbitMQ publisher confirm。只有重试消息或死信消息发布成功后，才会 `Ack` 原消息；如果发布失败，则 `Reject(true)` 保留原消息，避免在“消费失败后转发失败”的窗口丢消息。

### 7. RabbitMQ 异步削峰与 MySQL 幂等落库

RabbitMQ 消费者异步创建选课记录。消费端使用 MySQL 事务扣减真实库存，并通过学生 ID + 课程 ID 唯一索引保证重复消息不会重复落库。消费者落库成功后会将 `request:{studentID}:{courseID}` 从 `pending` 更新为 `success`，并写入选课结果缓存。库存不足会把请求状态更新为 `failed`，重复选课会按已成功结果处理。

### 8. 一致性取舍说明

Redis Stream Outbox 主要解决应用进程崩溃导致的消息丢失问题。Redis 自身故障时，可靠性取决于持久化策略。本项目在 Docker Compose 中开启 AOF，并使用 `appendfsync everysec`，在性能和可靠性之间做折中；理论上极端宕机场景仍可能丢失约 1 秒内尚未刷盘的数据。如果业务要求强一致不丢消息，可以改为 MySQL Outbox：在同一个 MySQL 事务中写业务表和消息表，再由后台任务投递 MQ。

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
