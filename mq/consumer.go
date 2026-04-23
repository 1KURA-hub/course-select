package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"go-course/global"
	"go-course/service"
	"strconv"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func Consumer() {
	// 获取消息管道
	msgs, err := global.MQChannel.Consume(
		global.Settings.RabbitMQ.QueueName,
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
				// 将单次消息处理抽离为独立函数 防止协程意外退出和内存泄漏
				processSingleMessage(d)
			}
		}(i)
	}
}

// 抽取出的单条消息处理逻辑
func processSingleMessage(d amqp091.Delivery) {
	global.Logger.Info("收到MQ消息", zap.String("msgID:", d.MessageId))
	var msg service.Message

	//  反序列化消息body
	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		global.Logger.Error("消息解析失败",
			zap.String("body", string(d.Body)),
			zap.Error(err))
		publishToDLQAndAck(d, "消息体反序列化失败")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 独立函数配合 defer，完美解决 Context 内存泄漏问题

	key := fmt.Sprintf("res:%d:%d", msg.StudentID, msg.CourseID)
	msgkey := fmt.Sprintf("msg:%d:%d", msg.StudentID, msg.CourseID)
	requestkey := fmt.Sprintf("request:%d:%d", msg.StudentID, msg.CourseID)

	// Redis分布式锁保证消息不重复
	exists := global.RDB.SetNX(ctx, msgkey, "processing", time.Second*10)
	if !exists.Val() {
		// 抢锁失败 确认锁的value
		status, _ := global.RDB.Get(ctx, msgkey).Result()
		if status == "success" {
			// 别人已经成功落库了 确认删除消息
			d.Ack(false)
			return
		}
		// processing说明别的消费者正在处理，转入延迟重试，避免立刻requeue造成空转。
		handleRetryOrDLQ(d, "消息正在处理中")
		return
	}

	// 数据库创建选课记录封装到了一个事务里面 并对TotalStock库存加锁 有错误则回滚
	err = service.CreateRecord(ctx, msg.StudentID, msg.CourseID)
	if err != nil {

		if err == service.ErrStockEmpty {
			if global.Logger.Core().Enabled(zap.DebugLevel) {
				global.Logger.Debug("库存不足",
					zap.Uint("课程ID", msg.CourseID))
			}
			// 消息消费成功 更新分布式锁value
			global.RDB.Set(ctx, msgkey, "success", time.Hour)
			global.RDB.Set(ctx, requestkey, "failed", 24*time.Hour)
			global.RDB.Set(ctx, key, -1, time.Minute)
			// 库存不足时 删除消息 进入下一次循环
			d.Ack(false)
			return
		}

		// 重复选课 两个消费者同时收到消息导致
		if err == service.ErrRepeatSelection {
			global.Logger.Warn("并发重复选课，已拦截",
				zap.Uint("uid", msg.StudentID),
				zap.Uint("cid", msg.CourseID))
			// 重复选课说明MySQL已经有最终成功记录，更新请求状态和结果缓存
			global.RDB.Set(ctx, msgkey, "success", time.Hour)
			global.RDB.Set(ctx, requestkey, "success", 24*time.Hour)
			global.RDB.Set(ctx, key, 1, time.Minute)
			// 确认消费 不要重试了
			d.Ack(false)
			return
		}

		global.Logger.Error("系统异常 创建记录失败",
			zap.Uint("学生ID", msg.StudentID),
			zap.Uint("课程ID", msg.CourseID),
			zap.Error(err))
		// 消息消费失败需要重试 删除分布式锁，避免重试消息被processing状态长期拦住。
		global.RDB.Del(ctx, msgkey)
		handleRetryOrDLQ(d, "创建选课记录失败")
		return
	}

	// 消息消费成功 数据库正常扣减
	global.RDB.Set(ctx, msgkey, "success", time.Hour)
	global.RDB.Set(ctx, requestkey, "success", 24*time.Hour)

	err = global.RDB.Set(ctx, key, 1, 5*time.Minute).Err()
	// 不能用defer 因为这里的for range是一个死循环 完成一次Redis操作直接cancel
	if err != nil {
		global.Logger.Error("Redis出错", zap.String("key", key), zap.Error(err))
		d.Ack(false)
		return
	}

	global.Logger.Info("选课记录创建成功",
		zap.Uint("学生ID", msg.StudentID),
		zap.Uint("课程ID", msg.CourseID))
	d.Ack(false)
}

func handleRetryOrDLQ(d amqp091.Delivery, reason string) {
	currentRetryCount := getRetryCount(d.Headers)
	nextRetryCount := currentRetryCount + 1
	routingKey, retrying := retryRoutingKey(nextRetryCount)
	if !retrying {
		routingKey = service.DLQRoutingKey
	}

	headers := copyHeaders(d.Headers)
	headers[service.RetryCountHeader] = nextRetryCount
	headers[service.FailedReasonHeader] = reason
	headers[service.FailedAtHeader] = time.Now().Format(time.RFC3339)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	publishing := clonePublishing(d, headers)
	if err := service.PublishWithConfirm(ctx, service.CourseSelectExchange, routingKey, publishing); err != nil {
		global.Logger.Error("失败消息转发失败，保留原消息等待RabbitMQ重新投递",
			zap.String("routingKey", routingKey),
			zap.Int("retryCount", nextRetryCount),
			zap.String("reason", reason),
			zap.Error(err))
		d.Reject(true)
		return
	}

	if retrying {
		global.Logger.Warn("消息处理失败，已转入延迟重试队列",
			zap.String("routingKey", routingKey),
			zap.Int("retryCount", nextRetryCount),
			zap.String("reason", reason))
	} else {
		global.Logger.Error("消息多次重试失败，已转入死信队列",
			zap.Int("retryCount", nextRetryCount),
			zap.String("reason", reason))
	}
	d.Ack(false)
}

func publishToDLQAndAck(d amqp091.Delivery, reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	headers := copyHeaders(d.Headers)
	headers[service.FailedReasonHeader] = reason
	headers[service.FailedAtHeader] = time.Now().Format(time.RFC3339)

	if err := service.PublishWithConfirm(ctx, service.CourseSelectExchange, service.DLQRoutingKey, clonePublishing(d, headers)); err != nil {
		global.Logger.Error("死信消息转发失败，保留原消息等待RabbitMQ重新投递", zap.Error(err))
		d.Reject(true)
		return
	}
	d.Ack(false)
}

func retryRoutingKey(retryCount int) (string, bool) {
	switch retryCount {
	case 1:
		return service.Retry1sRoutingKey, true
	case 2:
		return service.Retry5sRoutingKey, true
	case 3:
		return service.Retry10sRoutingKey, true
	default:
		return service.DLQRoutingKey, false
	}
}

func clonePublishing(d amqp091.Delivery, headers amqp091.Table) amqp091.Publishing {
	contentType := d.ContentType
	if contentType == "" {
		contentType = "application/json"
	}

	return amqp091.Publishing{
		Headers:         headers,
		ContentType:     contentType,
		ContentEncoding: d.ContentEncoding,
		DeliveryMode:    amqp091.Persistent,
		Priority:        d.Priority,
		CorrelationId:   d.CorrelationId,
		ReplyTo:         d.ReplyTo,
		Expiration:      d.Expiration,
		MessageId:       d.MessageId,
		Timestamp:       time.Now(),
		Type:            d.Type,
		UserId:          d.UserId,
		AppId:           d.AppId,
		Body:            d.Body,
	}
}

func copyHeaders(headers amqp091.Table) amqp091.Table {
	copied := amqp091.Table{}
	for key, value := range headers {
		copied[key] = value
	}
	return copied
}

func getRetryCount(headers amqp091.Table) int {
	value, ok := headers[service.RetryCountHeader]
	if !ok {
		return 0
	}

	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case string:
		count, err := strconv.Atoi(v)
		if err == nil {
			return count
		}
	}
	return 0
}
