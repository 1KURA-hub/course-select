# 高并发秒杀选课系统 (High-Concurrency Course Selection System)

## 项目简介

本项目是一个基于 Golang 开发的高并发秒杀选课系统，模拟高校选课/电商抢购场景。针对瞬时极高并发流量容易导致的“数据库宕机”、“缓存击穿”、“库存超卖”等痛点，系统采用**缓存前置、异步削峰**的微服务级架构，单机环境下成功实现 **1.5w+ TPS** 的吞吐量，且达成 **0 超卖、0 宕机**。

## 核心架构与技术亮点

### 1. 极限防超卖与接口防刷 (Redis Lua + 分布式锁)
摒弃传统的 MySQL 事务排队查库方案，基于 Redis `SetNX` 实现毫秒级用户防重连击拦截。核心扣减逻辑封装入**Redis Lua 脚本**，利用单线程特性在内存层面实现“库存校验+扣减+防重打标”的绝对原子性，完美杜绝并发超卖现象。

### 2. 极速缓存穿透拦截 (API 中间件 Bloom Filter)
系统启动阶段，全量预热课程 ID 至布隆过滤器。深度结合 Gin 路由的 RESTful 规范提取路径参数，**在中间件层前置拦截**恶意伪造请求，实现到达核心业务层之前的极速熔断与**零对象内存分配**。

### 3. 缓存击穿绝对防御 (Singleflight 机制)
针对热点课程缓存失效瞬间爆发的并发回源洪峰，底层引入 Go 官方扩展库 `golang.org/x/sync/singleflight`。利用 WaitGroup 的物理阻塞机制，将数万个相同的数据库查询请求合并为 1 个，极大保护了底层数据库免受瞬时击穿。

### 4. 异步削峰与最终一致性 (RabbitMQ Worker Pool)
对于抢课成功的流量，系统通过 RabbitMQ 异步投递落库消息。消费端采用 **Worker Pool 协程池**平缓拉取，并**关闭自动 ACK**；结合 `success/processing` 缓存状态机精细化控制 NACK 重试机制。底层配合 **MySQL 悲观锁 (`FOR UPDATE`)** 与唯一索引兜底，实现流量平滑过渡与绝对的幂等落库。

### 5. 全链路资源防泄露 (Context 级联控制)
应用 Go 并发哲学，将网关层 HTTP 请求的生命周期 (`c.Request.Context()`) 一路透传至 Redis I/O 与 GORM 底层，构建**父子级联超时树**。彻底解决因客户端异常断网或服务拥塞引发的后端“僵尸协程”堆积与内存泄漏问题。

## 性能压测报告 (JMeter)

在单机环境下，模拟 **1000 并发**，共计 **50,000 样本**的高并发攻击：

* **极致吞吐量 (Throughput):** 12010.6/sec (1.2w+ TPS)
* **平均响应时间 (Average RT):** 72 ms
* **错误率 (Error Rate):** 0.00% (达成 0 超卖、0 宕机)

> **压测结果截图：**  

![JMeter 压测报告](./docs/jmeter.png) 

## 快速启动 (Quick Start)

### 1. 环境依赖

* Golang >= 1.21
* MySQL >= 8.0
* Redis >= 7.0
* RabbitMQ >= 3.12

### 2. 初始化项目

您可以直接通过源码克隆并安装依赖：

```bash
# 克隆项目 (请将链接替换为你自己的 GitHub 仓库地址)
git clone [https://github.com/你的用户名/你的项目名.git](https://github.com/你的用户名/你的项目名.git)
cd 你的项目名

# 安装依赖
go mod tidy
```

### 3. 配置环境

运行前，请务必修改 `config.yaml` 或 `.env` 文件中的中间件连接信息：

```yaml
# 示例配置说明
MySQL: "root:123456@tcp(127.0.0.1:3306)/course_select"
Redis: "127.0.0.1:6379"
RabbitMQ: "amqp://guest:guest@127.0.0.1:5672/"
```

### 4. 运行服务

```bash
go run main.go
```

## 许可证

本项目采用 [MIT License](https://opensource.org/licenses/MIT) 开源许可证。