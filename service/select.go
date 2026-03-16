package service

import (
	"context"
	"fmt"
	"go-course/global"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 全局预加载 Lua 脚本，避免每次请求都重新编译，压榨极限性能
var selectScript = redis.NewScript(`
-- 1. 校验 预选课记录 是否存在
local exists = redis.call('get', KEYS[1])
if exists then
    return -1 -- ErrRepeatRequest
end

-- 2. 校验库存并扣减
local stock = redis.call('get', KEYS[2])
if not stock or tonumber(stock) < tonumber(ARGV[1]) then
    return 0 -- ErrStockEmpty
end

redis.call('decrby', KEYS[2], tonumber(ARGV[1]))
redis.call('set',KEYS[1],1,'EX',3600)
return 1 -- 成功
`)

func SelectCourse(studentID, courseID uint) error {
	// 布隆过滤器 快速过滤不存在的课程ID
	if !global.CourseBloomFilter.TestString(fmt.Sprintf("%d", courseID)) {
		return gorm.ErrRecordNotFound
	}

	// ctx超时控制 协程生命周期最多2s
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Redis分布式锁过滤短时间内大量重复请求
	recordkey := fmt.Sprintf("lock:%d:%d", studentID, courseID)

	repeat := global.RDB.SetNX(ctx, recordkey, studentID, 2*time.Second)

	// 判断是否加锁成功
	if !repeat.Val() {
		return ErrRepeatSelection
	}

	// 选课路径key和课程库存key
	requestkey := fmt.Sprintf("record:%d:%d", studentID, courseID)
	stockkey := fmt.Sprintf("course:stock:%d", courseID)

	// Lua keys args
	keys := []string{requestkey, stockkey}
	args := []interface{}{1}

	// Lua脚本实现不超卖 查询库存和扣减库存成为原子性操作
	res, err := selectScript.Run(ctx, global.RDB, keys, args...).Int()
	if err != nil {
		global.Logger.Error("Lua脚本执行出错", zap.Error(err))
		return ErrSystemBusy
	}

	// 判断Lua脚本执行结果
	switch res {
	case -1:
		if global.Logger.Core().Enabled(zap.DebugLevel) {
			global.Logger.Debug("用户重复发送请求", zap.Uint("studentID", studentID))
		}
		return ErrRepeatRequest
	case 0:
		if global.Logger.Core().Enabled(zap.DebugLevel) {
			global.Logger.Debug("库存不足", zap.Uint("courseID", courseID))
		}
		return ErrStockEmpty
	case 1:
	default:
		global.Logger.Error("Lua 脚本返回了未知的状态码", zap.Int("res", res))
		return ErrSystemBusy
	}

	// 预扣减成功 发送消息到MQ
	err = Send(studentID, courseID)
	if err != nil {
		global.RDB.Incr(ctx, stockkey)
		global.RDB.Del(ctx, requestkey)
		global.Logger.Error("消息发送出错", zap.Error(err))
		return ErrSystemBusy
	}

	global.Logger.Info("消息发送成功", zap.Uint("studentID", studentID), zap.Uint("courseID", courseID))
	return nil
}
