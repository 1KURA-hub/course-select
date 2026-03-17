package initialize

import (
	"go-course/global"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// 初始化RabbitMQ连接
func InitRabbitMQ() {
	// 建立服务器与MQ的连接
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		global.Logger.Fatal("连接RabbitMQ失败", zap.Error(err))
	}

	// 声明通道
	global.MQChannel, err = conn.Channel()
	if err != nil {
		global.Logger.Fatal("通道声明失败", zap.Error(err))
		panic(err)
	}

	// 声明队列
	_, err = global.MQChannel.QueueDeclare(
		"redisQueue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		global.Logger.Fatal("队伍声明失败", zap.Error(err))
	}
	global.Logger.Info("初始化RabbitMQ成功")
}
