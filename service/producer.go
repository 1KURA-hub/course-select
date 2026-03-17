package service

import (
	"context"
	"encoding/json"
	"errors"
	"go-course/global"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Message struct {
	StudentID uint
	CourseID  uint
}

// 发送消息到MQ
func Send(studentID, courseID uint) error {
	var message = Message{
		StudentID: studentID,
		CourseID:  courseID,
	}

	// message序列化
	body, err := json.Marshal(message)
	if err != nil {
		global.Logger.Error("消息序列化失败", zap.Error(err))
		return errors.New("系统原因 选课失败")
	}

	// 防止网络抖动导致生产者卡死
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// MQ通道发布消息
	err = global.MQChannel.PublishWithContext(
		ctx,
		"",
		"redisQueue",
		false,
		false,
		amqp.Publishing{
			// 告诉消费者 我发的是JSON
			ContentType: "application/json",
			// 消息持久化模式 确保MQ重启后消息不丢失
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)

	if err != nil {
		global.Logger.Error("消息发布失败", zap.Error(err))

		return errors.New("系统原因 选课失败")
	}
	return nil
}
