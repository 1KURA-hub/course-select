package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"go-course/global"

	"go-course/service"
	"time"

	"go.uber.org/zap"
)

func Consumer() {
	// 获取消息管道
	msgs, err := global.MQChannel.Consume(
		"redisQueue",
		"",    // consumer
		false, // auto-ack: 必须关闭自动确认 手动保证数据不丢失
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		global.Logger.Fatal("MQ消费者启动失败", zap.Error(err))
		return
	}

	WorkerNum := 10

	for i := 0; i < WorkerNum; i++ {
		// 如果不传参数 闭包捕获循环变量 i 相当于引用同一个地址 最后i相同
		go func(workerID int) {
			global.Logger.Info("消费者Worker启动", zap.Int("workerID", workerID))

			for d := range msgs {
				global.Logger.Info("收到MQ消息", zap.String("msgID:", d.MessageId))
				var msg service.Message
				//  反序列化消息body
				err = json.Unmarshal(d.Body, &msg)
				if err != nil {
					global.Logger.Error("消息解析失败",
						zap.String("body", string(d.Body)),
						zap.Error(err))
					//消息解析失败 删除这条消息 进行下一次循环
					d.Ack(false)
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				key := fmt.Sprintf("res:%d:%d", msg.StudentID, msg.CourseID)

				// 数据库创建选课记录封装到了一个事务里面 并对TotalStock库存加锁 有错误则回滚
				err = service.CreateRecord(msg.StudentID, msg.CourseID)
				if err != nil {

					if err == service.ErrStockEmpty {
						if global.Logger.Core().Enabled(zap.DebugLevel) {
							global.Logger.Debug("库存不足",
								zap.Uint("课程ID", msg.CourseID))
						}
						global.RDB.Set(ctx, key, -1, time.Minute)
						cancel()
						// 库存不足时 删除消息 进入下一次循环
						d.Ack(false)
						continue
					}
					// 重复选课 两个消费者同时收到消息导致
					if err == service.ErrRepeatSelection {
						global.Logger.Warn("并发重复选课，已拦截",
							zap.Uint("uid", msg.StudentID),
							zap.Uint("cid", msg.CourseID))
						global.RDB.Set(ctx, key, -1, time.Minute)
						cancel()
						d.Ack(false) // 确认消费 不要重试了
						continue
					}
					global.Logger.Error("系统异常 创建记录失败",
						zap.Uint("学生ID", msg.StudentID),
						zap.Uint("课程ID", msg.CourseID),
						zap.Error(err))
					time.Sleep(1 * time.Second)
					cancel()
					// 系统出错 把这条消息重新加入消息队列
					d.Reject(true)
					continue

				}

				err = global.RDB.Set(ctx, key, 1, 5*time.Minute).Err()
				// 不能用defer 因为这里的for range是一个死循环 完成一次Redis操作直接cancel
				cancel()
				if err != nil {
					global.Logger.Error("Redis出错", zap.String("key", key), zap.Error(err))
					d.Ack(false)
					continue
				}

				global.Logger.Info("选课记录创建成功",
					zap.Uint("学生ID", msg.StudentID),
					zap.Uint("课程ID", msg.CourseID))
				d.Ack(false)
			}
		}(i)
	}
}
