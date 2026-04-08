#  高并发秒杀选课系统 (High-Concurrency Course Selection)

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Redis](https://img.shields.io/badge/Redis-7.0+-DC382D?style=flat&logo=redis)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.12+-FF6600?style=flat&logo=rabbitmq)
![MySQL](https://img.shields.io/badge/MySQL-8.0+-4479A1?style=flat&logo=mysql)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

## 📖 项目简介

本项目是一个基于 Golang 构建的高并发秒杀/选课系统，模拟高校海量学生同时抢课的极端瞬时并发场景。系统旨在设计并实现高并发场景下的安全、缓存与数据一致性方案。

**技术栈：** Golang, Gin, GORM, MySQL, Redis, RabbitMQ, Zap, Docker

## 🚀 核心架构与技术亮点

### 1. 安全与限流 (网关拦截)
基于 Gin 构建 API 网关，集成 JWT 实现无状态鉴权。核心接口利用布隆过滤器中间件，成功拦截 90% 以上无效流量，稳控服务入口。

### 2. 缓存防护 (防穿透与击穿)
采用“布隆过滤器+缓存空值”解决缓存穿透问题；结合扩展库 `singleflight` 合并并发重复查询为一次网络请求，有效防止缓存击穿，大幅降低数据库读取压力。

### 3. 数据一致性 (0 超卖保障)
使用 Redis Lua 脚本原子化实现请求去重与库存预扣减，保障高并发场景下 0 超卖。消费端依靠分布式锁与 MySQL 唯一索引拦截重复消息，保证最终的幂等性。

### 4. 异步削峰解耦 (吞吐量提升)
引入 RabbitMQ 异步处理选课写入，削峰填谷承接瞬时高并发流量。相较同步直写，系统吞吐量提升 6 倍，有效避免数据库宕机。

## 📊 部署与性能压测

基于 Docker 容器化部署至 2C4G 云服务器，通过 pprof 定位并优化网络 RTT 瓶颈。
使用 **wrk** 实测 1 万独立 Token 争抢 1000 库存（200 并发）：

* **峰值吞吐量 (QPS):** 6132 req/sec
* **平均延迟 (Average RT):** 32.52 ms
* **核心指标:** 实现了异步落盘和 0 超卖，平稳无宕机，验证了系统高并发稳定性。

##  快速启动 (Quick Start)

### 1. 环境依赖
* Golang >= 1.21
* MySQL >= 8.0
* Redis >= 7.0
* RabbitMQ >= 3.12

### 2. 克隆与运行
```bash
# 1. 克隆项目
git clone [https://github.com/1KURA-hub/course-select.git](https://github.com/1KURA-hub/course-select.git)
cd course-select

# 2. 安装依赖
go mod tidy

# 3. 配置环境 (请修改 config.yaml 中的中间件连接地址)
# MySQL: root:123456@tcp(127.0.0.1:3306)/course_select
# Redis: 127.0.0.1:6379
# RabbitMQ: amqp://guest:guest@127.0.0.1:5672/

# 4. 启动服务
go run main.go
```
## 许可证
本项目采用 [MIT License](https://opensource.org/licenses/MIT) 开源许可证。