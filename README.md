#  高并发秒杀选课系统 (High-Concurrency Course Selection)

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Redis](https://img.shields.io/badge/Redis-7.0+-DC382D?style=flat&logo=redis)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.12+-FF6600?style=flat&logo=rabbitmq)
![MySQL](https://img.shields.io/badge/MySQL-8.0+-4479A1?style=flat&logo=mysql)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

##  项目简介

本项目是一个基于 Golang 构建的高并发秒杀/选课系统，模拟高校海量学生同时抢课的极端瞬时并发场景。

系统摒弃了传统的“数据库抗压”思维，创新性地采用 **“漏斗型拦截架构 (Funnel Architecture)”**。通过网关层、缓存层、消息队列层的层层过滤与异步削峰，成功将海量无效的读写并发拦截在数据库之外。在单机压测环境下，系统成功承载 **1.2w 峰值 QPS**，且达成核心链路 **0 超卖、0 宕机、0 消息丢失**。

##  核心架构与技术亮点

> **💡 架构设计核心思想：Fail-Fast (快速失败) 与 内存级前置拦截**

### 1. 多级防线与漏斗拦截 (JWT + Bloom Filter + SETNX)
在流量触达核心业务前构建三道防线：
* **路由网关层**：无状态 JWT 校验身份，**布隆过滤器 (Bloom Filter)** 极速拦截伪造的非法课程 ID，实现零对象内存分配。
* **操作防重层**：基于 Redis `SETNX` 构建短 TTL 分布式锁，精准拦截用户手抖连击与恶意重放脚本。
* **效果**：通过前置防线，将 **90% 以上** 的无效或恶意流量拦截在底层 DB 之外。

### 2. 内存级原子扣减 (Redis Lua)
摒弃传统的 MySQL 事务排队扣减方案。将“库存校验 + 扣减 + 用户去重打标”三大逻辑封装进 **Redis Lua 脚本**。利用 Redis 的单线程排他性，在纯内存级别完成毫秒级原子操作，从根源上杜绝并发超卖现象。

### 3. 缓存击穿并发风暴防御 (Singleflight)
针对热点课程缓存失效（或预热前）瞬间爆发的并发回源洪峰，底层引入 Go 官方扩展库 `golang.org/x/sync/singleflight`。在应用级内存中将数万个相同的数据库查询请求合并为 1 个，实现**零网络开销**的 DB 并发风暴防御。

### 4. 异步削峰与最终一致性 (RabbitMQ + 唯一索引)
对内存扣减成功的极少量有效流量，通过 RabbitMQ 异步投递落库消息，将随机的写入洪峰转化为平滑的匀速落库（削峰填谷）。
* **可靠性兜底**：消费端开启**严格手动 ACK**。底层彻底放弃容易引发死锁的行级悲观锁，改用 MySQL **唯一联合索引 (User_ID + Course_ID)** 作为最终防线，配合 Redis 分布式锁实现消费端的绝对幂等，保证 0 消息丢失与 0 重复消费。

### 5. 链路资源防泄露 (Context 级联控制)
应用 Go 并发哲学，将网关层 HTTP 请求的生命周期 (`c.Request.Context()`) 一路透传至 Redis I/O 与 GORM 底层，构建**父子级联超时树**。彻底解决因客户端断网或下游服务拥塞引发的 Goroutine 堆积与内存泄漏问题。

##  性能压测报告 (JMeter)

在普通单机环境下，使用 JMeter 模拟 **1000 线程并发**，共计 **50,000 样本**的瞬时洪峰攻击：

* **峰值吞吐量 (Throughput/QPS):** 12010.6 req/sec
* **核心链路平均响应时间 (Average RT):** 72 ms
* **错误率 (Error Rate):** 0.00% (完美达成 0 超卖、系统平稳无宕机)

![JMeter 压测报告](./docs/jmeter.png)

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