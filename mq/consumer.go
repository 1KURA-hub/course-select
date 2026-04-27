package mq

import (
	"context"
	"encoding/json"
	"go-course/global"
	redisrepo "go-course/repository/redis"
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
		go func(workerID int) {
			global.Logger.Info("消费者Worker启动", zap.Int("workerID", workerID))

			for d := range msgs {
				processSingleMessage(d)
			}
		}(i)
	}
}

func processSingleMessage(d amqp091.Delivery) {
	global.Logger.Info("收到MQ消息", zap.String("msgID", d.MessageId))
	var msg Message

	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		global.Logger.Error("消息解析失败",
			zap.String("body", string(d.Body)),
			zap.Error(err))
		publishToDLQAndAck(d, "消息体反序列化失败")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = service.CreateRecord(ctx, msg.StudentID, msg.CourseID)
	if err != nil {
		if err == service.ErrStockEmpty {
			if global.Logger.Core().Enabled(zap.DebugLevel) {
				global.Logger.Debug("库存不足", zap.Uint("课程ID", msg.CourseID))
			}
			if cacheErr := redisrepo.MarkSelectionFailed(ctx, msg.StudentID, msg.CourseID); cacheErr != nil {
				global.Logger.Warn("库存不足结果写入Redis失败，按最终失败直接确认消息",
					zap.Uint("studentID", msg.StudentID),
					zap.Uint("courseID", msg.CourseID),
					zap.Error(cacheErr))
			}
			d.Ack(false)
			return
		}

		if err == service.ErrRepeatSelection {
			global.Logger.Warn("并发重复选课，已拦截",
				zap.Uint("uid", msg.StudentID),
				zap.Uint("cid", msg.CourseID))
			if cacheErr := redisrepo.MarkSelectionSuccess(ctx, msg.StudentID, msg.CourseID); cacheErr != nil {
				global.Logger.Warn("重复选课成功结果写入Redis失败，依赖MySQL最终事实兜底",
					zap.Uint("studentID", msg.StudentID),
					zap.Uint("courseID", msg.CourseID),
					zap.Error(cacheErr))
			}
			d.Ack(false)
			return
		}

		global.Logger.Error("系统异常 创建记录失败",
			zap.Uint("学生ID", msg.StudentID),
			zap.Uint("课程ID", msg.CourseID),
			zap.Error(err))
		handleRetryOrDLQ(d, "创建选课记录失败")
		return
	}

	if cacheErr := redisrepo.MarkSelectionSuccess(ctx, msg.StudentID, msg.CourseID); cacheErr != nil {
		global.Logger.Warn("选课成功结果写入Redis失败，依赖MySQL最终事实兜底",
			zap.Uint("studentID", msg.StudentID),
			zap.Uint("courseID", msg.CourseID),
			zap.Error(cacheErr))
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
		routingKey = DLQRoutingKey
	}

	headers := copyHeaders(d.Headers)
	headers[RetryCountHeader] = nextRetryCount
	headers[FailedReasonHeader] = reason
	headers[FailedAtHeader] = time.Now().Format(time.RFC3339)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	publishing := clonePublishing(d, headers)
	if err := PublishWithConfirm(ctx, CourseSelectExchange, routingKey, publishing); err != nil {
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
	headers[FailedReasonHeader] = reason
	headers[FailedAtHeader] = time.Now().Format(time.RFC3339)

	if err := PublishWithConfirm(ctx, CourseSelectExchange, DLQRoutingKey, clonePublishing(d, headers)); err != nil {
		global.Logger.Error("死信消息转发失败，保留原消息等待RabbitMQ重新投递", zap.Error(err))
		d.Reject(true)
		return
	}
	d.Ack(false)
}

func retryRoutingKey(retryCount int) (string, bool) {
	switch retryCount {
	case 1:
		return Retry1sRoutingKey, true
	case 2:
		return Retry5sRoutingKey, true
	case 3:
		return Retry10sRoutingKey, true
	default:
		return DLQRoutingKey, false
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
	value, ok := headers[RetryCountHeader]
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
