package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-course/global"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Message struct {
	StudentID uint
	CourseID  uint
}

// Send 发送消息到MQ，保留给需要直接投递的调用方。
func Send(studentID, courseID uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return SendWithContext(ctx, studentID, courseID)
}

// SendWithContext 发送消息到MQ，并等待 RabbitMQ publisher confirm。
func SendWithContext(ctx context.Context, studentID, courseID uint) error {
	var message = Message{
		StudentID: studentID,
		CourseID:  courseID,
	}

	body, err := json.Marshal(message)
	if err != nil {
		global.Logger.Error("消息序列化失败", zap.Error(err))
		return errors.New("系统原因 选课失败")
	}

	if global.MQPublishChannel == nil || global.MQConfirmChan == nil {
		global.Logger.Error("MQ发布通道未初始化")
		return errors.New("系统原因 选课失败")
	}

	err = global.MQPublishChannel.PublishWithContext(
		ctx,
		"",
		global.Settings.RabbitMQ.QueueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    fmt.Sprintf("%d:%d:%d", studentID, courseID, time.Now().UnixNano()),
			Body:         body,
		},
	)
	if err != nil {
		global.Logger.Error("消息发布失败", zap.Error(err))
		return errors.New("系统原因 选课失败")
	}

	select {
	case confirm, ok := <-global.MQConfirmChan:
		if !ok {
			global.Logger.Error("MQ发布确认通道已关闭")
			return errors.New("系统原因 选课失败")
		}
		if !confirm.Ack {
			global.Logger.Error("MQ消息发布未被确认", zap.Uint64("deliveryTag", confirm.DeliveryTag))
			return errors.New("系统原因 选课失败")
		}
		return nil
	case <-ctx.Done():
		global.Logger.Error("等待MQ发布确认超时", zap.Error(ctx.Err()))
		return errors.New("系统原因 选课失败")
	}
}
