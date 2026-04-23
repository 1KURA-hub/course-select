package initialize

import (
	"fmt"
	"go-course/global"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// 初始化RabbitMQ连接
func InitRabbitMQ() {
	mqConf := global.Settings.RabbitMQ
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		mqConf.User,
		mqConf.Password,
		mqConf.Host,
		mqConf.Port,
	)

	conn, err := amqp.Dial(dsn)
	if err != nil {
		global.Logger.Fatal("连接RabbitMQ失败", zap.Error(err))
	}
	global.MQConn = conn

	global.MQChannel, err = conn.Channel()
	if err != nil {
		global.Logger.Fatal("消费通道声明失败", zap.Error(err))
	}

	_, err = global.MQChannel.QueueDeclare(
		mqConf.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		global.Logger.Fatal("队列声明失败", zap.Error(err))
	}

	global.MQPublishChannel, err = conn.Channel()
	if err != nil {
		global.Logger.Fatal("发布通道声明失败", zap.Error(err))
	}

	if err = global.MQPublishChannel.Confirm(false); err != nil {
		global.Logger.Fatal("发布确认模式开启失败", zap.Error(err))
	}
	global.MQConfirmChan = global.MQPublishChannel.NotifyPublish(make(chan amqp.Confirmation, 1))

	global.Logger.Info("初始化RabbitMQ成功")
}
