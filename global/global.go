package global

import (
	"go-course/config"
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 全局配置变量(MySQL Redis Log)
var Settings config.Config

// 全局数据库连接对象
var DB *gorm.DB

// 全局Redis连接对象
var RDB *redis.Client

// 全局日志变量
var Logger *zap.Logger

// 全局MQ连接和消费通道
var MQConn *amqp.Connection
var MQChannel *amqp.Channel

// 全局MQ发布通道和发布确认通道
var MQPublishChannel *amqp.Channel
var MQConfirmChan <-chan amqp.Confirmation
var MQPublishMu sync.Mutex

// 全局声明一个布隆过滤器指针
var CourseBloomFilter *bloom.BloomFilter
