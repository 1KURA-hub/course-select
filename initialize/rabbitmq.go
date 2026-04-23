package initialize

import (
	"fmt"
	"go-course/global"
	"go-course/service"

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

	if err = declareCourseSelectTopology(global.MQChannel, mqConf.QueueName); err != nil {
		global.Logger.Fatal("RabbitMQ选课队列拓扑声明失败", zap.Error(err))
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

func declareCourseSelectTopology(ch *amqp.Channel, mainQueue string) error {
	if err := ch.ExchangeDeclare(
		service.CourseSelectExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if err := declareAndBindQueue(ch, mainQueue, service.MainRoutingKey, nil); err != nil {
		return err
	}

	retryQueues := []struct {
		name       string
		routingKey string
		ttl        int32
	}{
		{name: mainQueue + ".retry.1s", routingKey: service.Retry1sRoutingKey, ttl: 1000},
		{name: mainQueue + ".retry.5s", routingKey: service.Retry5sRoutingKey, ttl: 5000},
		{name: mainQueue + ".retry.10s", routingKey: service.Retry10sRoutingKey, ttl: 10000},
	}

	for _, queue := range retryQueues {
		args := amqp.Table{
			"x-message-ttl":             queue.ttl,
			"x-dead-letter-exchange":    service.CourseSelectExchange,
			"x-dead-letter-routing-key": service.MainRoutingKey,
		}
		if err := declareAndBindQueue(ch, queue.name, queue.routingKey, args); err != nil {
			return err
		}
	}

	return declareAndBindQueue(ch, mainQueue+".dlq", service.DLQRoutingKey, nil)
}

func declareAndBindQueue(ch *amqp.Channel, queueName, routingKey string, args amqp.Table) error {
	_, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return err
	}

	return ch.QueueBind(
		queueName,
		routingKey,
		service.CourseSelectExchange,
		false,
		nil,
	)
}
