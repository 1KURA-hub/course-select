package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-course/global"
	"go-course/initialize"
	"go-course/model"
	"go-course/mq"
	mysqlrepo "go-course/repository/mysql"
	redisrepo "go-course/repository/redis"
	"go-course/service"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	initialize.InitConfig()
	initialize.InitLogger()
	initialize.InitMySQL()
	initialize.InitRedis()
	initialize.InitRabbitMQ()

	defer func() {
		if global.MQChannel != nil {
			_ = global.MQChannel.Close()
		}
		if global.MQPublishChannel != nil {
			_ = global.MQPublishChannel.Close()
		}
		if global.MQConn != nil {
			_ = global.MQConn.Close()
		}
	}()

	dlqName := global.Settings.RabbitMQ.QueueName + ".dlq"
	processed := 0

	for {
		d, ok, err := global.MQChannel.Get(dlqName, false)
		if err != nil {
			panic(fmt.Errorf("读取死信队列失败: %w", err))
		}
		if !ok {
			fmt.Printf("DLQ 已清空，本次共处理 %d 条消息\n", processed)
			return
		}

		if err = compensateOne(d); err != nil {
			_ = d.Nack(false, true)
			panic(fmt.Errorf("补偿死信消息失败: %w", err))
		}

		if err = d.Ack(false); err != nil {
			panic(fmt.Errorf("ACK 死信消息失败: %w", err))
		}
		processed++
	}
}

func compensateOne(d amqp091.Delivery) error {
	var msg mq.Message
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		global.Logger.Error("死信消息反序列化失败，直接丢弃", zap.Error(err))
		return nil
	}

	selection, err := mysqlrepo.GetSelectionBySIDAndCID(msg.StudentID, msg.CourseID)
	if err == nil && selection != nil && selection.Status == model.SelectionStatusSelected {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if cacheErr := redisrepo.MarkSelectionRequestSuccess(ctx, msg.StudentID, msg.CourseID); cacheErr != nil {
			global.Logger.Warn("死信补偿命中已落库记录，但Redis结果写回失败", zap.Error(cacheErr))
		}
		fmt.Printf("跳过已落库消息 student=%d course=%d\n", msg.StudentID, msg.CourseID)
		return nil
	}
	if err == nil && selection != nil && selection.Status == model.SelectionStatusDropped {
		fmt.Printf("跳过已退课消息 student=%d course=%d\n", msg.StudentID, msg.CourseID)
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = service.CompensateRecord(ctx, msg.StudentID, msg.CourseID)
	if err != nil && !errors.Is(err, service.ErrRepeatSelection) {
		return err
	}

	if cacheErr := redisrepo.MarkSelectionRequestSuccess(ctx, msg.StudentID, msg.CourseID); cacheErr != nil {
		global.Logger.Warn("死信补偿落库成功，但Redis结果写回失败", zap.Error(cacheErr))
	}

	fmt.Printf("补偿成功 student=%d course=%d\n", msg.StudentID, msg.CourseID)
	return nil
}
