package mq

import (
	"context"
	"errors"
	"fmt"
	"go-course/global"
	"go-course/service"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	selectRelayGroup   = "select-relay-group"
	relayReadBatchSize = 10
	relayBlockTime     = 5 * time.Second
	relayClaimIdleTime = 60 * time.Second
	relayClaimInterval = 30 * time.Second
)

// StartRelay 启动 Redis Stream -> RabbitMQ 的中继投递协程。
func StartRelay() {
	ctx := context.Background()
	if err := ensureRelayGroup(ctx); err != nil {
		global.Logger.Fatal("Redis Stream消费组初始化失败", zap.Error(err))
	}

	hostname, _ := os.Hostname()
	consumerName := fmt.Sprintf("relay-%s-%d", hostname, os.Getpid())

	go relayNewMessages(consumerName)
	go reclaimPendingMessages(consumerName)
}

func ensureRelayGroup(ctx context.Context) error {
	err := global.RDB.XGroupCreateMkStream(ctx, service.SelectStreamKey, selectRelayGroup, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

func relayNewMessages(consumerName string) {
	for {
		streams, err := global.RDB.XReadGroup(context.Background(), &redis.XReadGroupArgs{
			Group:    selectRelayGroup,
			Consumer: consumerName,
			Streams:  []string{service.SelectStreamKey, ">"},
			Count:    relayReadBatchSize,
			Block:    relayBlockTime,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			global.Logger.Error("读取Redis Stream新消息失败", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				handleStreamMessage(msg)
			}
		}
	}
}

func reclaimPendingMessages(consumerName string) {
	ticker := time.NewTicker(relayClaimInterval)
	defer ticker.Stop()

	for range ticker.C {
		start := "0-0"
		for {
			messages, next, err := global.RDB.XAutoClaim(context.Background(), &redis.XAutoClaimArgs{
				Stream:   service.SelectStreamKey,
				Group:    selectRelayGroup,
				Consumer: consumerName,
				MinIdle:  relayClaimIdleTime,
				Start:    start,
				Count:    relayReadBatchSize,
			}).Result()
			if err != nil {
				if !errors.Is(err, redis.Nil) {
					global.Logger.Error("回收Redis Stream pending消息失败", zap.Error(err))
				}
				break
			}

			if len(messages) == 0 {
				break
			}

			for _, msg := range messages {
				handleStreamMessage(msg)
			}

			if next == "0-0" || next == start {
				break
			}
			start = next
		}
	}
}

func handleStreamMessage(msg redis.XMessage) {
	studentID, err := streamUint(msg.Values["student_id"])
	if err != nil {
		global.Logger.Error("Redis Stream消息缺少学生ID", zap.String("messageID", msg.ID), zap.Error(err))
		return
	}

	courseID, err := streamUint(msg.Values["course_id"])
	if err != nil {
		global.Logger.Error("Redis Stream消息缺少课程ID", zap.String("messageID", msg.ID), zap.Error(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := service.SendWithContext(ctx, studentID, courseID); err != nil {
		global.Logger.Error("Redis Stream消息投递MQ失败",
			zap.String("messageID", msg.ID),
			zap.Uint("studentID", studentID),
			zap.Uint("courseID", courseID),
			zap.Error(err))
		return
	}

	if err := global.RDB.XAck(ctx, service.SelectStreamKey, selectRelayGroup, msg.ID).Err(); err != nil {
		global.Logger.Error("Redis Stream消息确认失败", zap.String("messageID", msg.ID), zap.Error(err))
		return
	}

	global.Logger.Info("Redis Stream消息已投递MQ",
		zap.String("messageID", msg.ID),
		zap.Uint("studentID", studentID),
		zap.Uint("courseID", courseID))
}

func streamUint(value interface{}) (uint, error) {
	var raw string
	switch v := value.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	case int64:
		return uint(v), nil
	case uint64:
		return uint(v), nil
	case int:
		return uint(v), nil
	default:
		raw = fmt.Sprint(v)
	}

	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}
