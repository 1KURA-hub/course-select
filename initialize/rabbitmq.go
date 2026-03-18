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
	// 2. 动态拼接 DSN (Data Source Name)
	// 格式: amqp://user:password@host:port/
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		mqConf.User,
		mqConf.Password,
		mqConf.Host,
		mqConf.Port,
	)
	// 建立服务器与MQ的连接
	conn, err := amqp.Dial(dsn)
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
		mqConf.QueueName,
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
