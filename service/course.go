package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-course/dao"
	"go-course/global"
	"go-course/model"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var sf singleflight.Group

func GetCourseList() ([]model.Course, error) {
	return dao.GetCourseList()
}

func GetCourseById(id uint) (*model.Course, error) {
	// 布隆过滤器验证ID是否存在
	if !global.CourseBloomFilter.TestString(fmt.Sprintf("%d", id)) {
		global.Logger.Warn("布隆过滤器拦截了恶意伪造的课程ID", zap.Uint("id", id))
		return nil, gorm.ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 缓存空值
	coursekey := fmt.Sprintf("course:%d", id)
	val, err := global.RDB.Get(ctx, coursekey).Result()

	if err == nil {
		// 查询命中空值
		if val == "" {
			global.Logger.Warn("拦截恶意请求 没有这个课程", zap.Uint("id:", id), zap.Error(err))
			return nil, gorm.ErrRecordNotFound
		}
		var course model.Course
		err = json.Unmarshal([]byte(val), &course)
		if err != nil {
			global.Logger.Error("反序列化course时出错", zap.Error(err))
			return nil, ErrSystemBusy
		}

		return &course, nil

	}
	// 未命中缓存
	if errors.Is(err, redis.Nil) {
		global.Logger.Debug("Redis查询为空")
		// 高并发下 保证同一时间只有一个协程查询数据库构建缓存
		v, sfErr, shared := sf.Do(coursekey, func() (interface{}, error) {
			course, dbErr := dao.GetCourseById(id)
			if dbErr == nil {
				// 结构体序列化为JSON存入Redis
				courseJSON, err := json.Marshal(course)
				if err != nil {
					global.Logger.Error("序列化course结构体出错", zap.Error(err))
					return nil, ErrSystemBusy
				}
				err = global.RDB.Set(ctx, coursekey, courseJSON, time.Minute).Err()
				if err != nil {
					global.Logger.Error("Redis出错", zap.Error(err))
					return nil, ErrSystemBusy
				}
				return course, nil
			}
			// 查询数据库为空 缓存空值
			if errors.Is(dbErr, gorm.ErrRecordNotFound) {
				global.Logger.Debug("数据库查询为空")
				rdbErr := global.RDB.Set(ctx, coursekey, "", time.Minute).Err()
				if rdbErr != nil {
					global.Logger.Error("Redis出错", zap.Error(rdbErr))
					return nil, ErrSystemBusy
				}
				return nil, gorm.ErrRecordNotFound
			}
			global.Logger.Error("数据库回源出错", zap.Error(dbErr))
			return nil, ErrSystemBusy
		})
		// singleflight内闭包函数结束
		if sfErr != nil {
			return nil, sfErr
		}
		// 记录被拦截并且共享结果的请求
		if shared {
			global.Logger.Debug("singleflight成功合并了并发请求", zap.Uint("id", id))
		}
		// 类型断言 还原成course结构体
		sfResult := v.(*model.Course)
		return sfResult, nil
	}
	global.Logger.Error("Redis出错", zap.Error(err))
	return nil, ErrSystemBusy
}
