package service

import (
	"context"
	"errors"
	"fmt"
	"go-course/global"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	SelectStreamKey      = "select:stream"
	requestKeyTTLSeconds = 86400
)

// 全局预加载 Lua 脚本 避免每次请求都重新编译
var selectScript = redis.NewScript(`
-- 1. 校验预选课请求是否存在
local exists = redis.call('get', KEYS[1])
if exists then
    return -1 -- ErrRepeatRequest
end

-- 2. 校验库存并扣减
local stock = redis.call('get', KEYS[2])
if not stock or tonumber(stock) < tonumber(ARGV[1]) then
    return 0 -- ErrStockEmpty
end

-- 3. 原子完成库存预扣、写入防重标记、写入 Redis Stream 消息表
redis.call('decrby', KEYS[2], tonumber(ARGV[1]))
redis.call('set', KEYS[1], '1', 'EX', tonumber(ARGV[4]))
redis.call('xadd', KEYS[3], '*', 'student_id', ARGV[2], 'course_id', ARGV[3])
return 1 -- 成功
`)

func SelectCourse(ctx context.Context, studentID, courseID uint) error {
	// 网关ctx级联取消 timeoutCtx超时控制 业务逻辑生命周期最多2s
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	requestkey := RequestKey(studentID, courseID)
	stockkey := fmt.Sprintf("course:stock:%d", courseID)

	keys := []string{requestkey, stockkey, SelectStreamKey}
	args := []interface{}{1, studentID, courseID, requestKeyTTLSeconds}

	res, err := selectScript.Run(timeoutCtx, global.RDB, keys, args...).Int()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return ErrSystemBusy
		}
		global.Logger.Error("Lua脚本执行出错", zap.Error(err))
		return ErrSystemBusy
	}

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

	global.Logger.Info("选课请求已写入Redis Stream", zap.Uint("studentID", studentID), zap.Uint("courseID", courseID))
	return nil
}
