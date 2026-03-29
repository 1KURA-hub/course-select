package initialize

import (
	"context"
	"fmt"
	"go-course/dao"
	"go-course/global"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// 初始化Redis连接 并预热课程库存
func InitRedis() {
	r := global.Settings.Redis
	global.RDB = redis.NewClient(&redis.Options{
		Addr:         r.Addr,
		Password:     r.Password,
		DB:           r.DB,
		PoolSize:     100,
		MinIdleConns: 10,
	})
	err := global.RDB.Ping(context.Background()).Err()
	if err != nil {
		global.Logger.Fatal("初始化Redis失败", zap.Error(err))
	}
	global.Logger.Info("初始化Redis成功")
	err = LoadStockToRedis()
	if err != nil {
		global.Logger.Fatal("Redis课程库存初始化失败", zap.Error(err))
	}
	global.Logger.Info("所有课程库存预热成功")
}

// 预热Redis课程库存函数
func LoadStockToRedis() error {
	// 协程超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 获取课程列表信息
	courses, err := dao.GetCourseList(ctx)
	if err != nil {
		return err
	}
	// 预热库存到Redis
	for _, course := range courses {
		Key := fmt.Sprintf("course:stock:%d", course.ID)
		err = global.RDB.Set(ctx, Key, course.Stock, 0).Err()
		if err != nil {
			return err
		}
		global.Logger.Info(fmt.Sprintf("课程%d预热成功,库存为%d", course.ID, course.Stock))
	}
	return nil
}
