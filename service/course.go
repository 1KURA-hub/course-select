package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-course/global"
	"go-course/model"
	mysqlrepo "go-course/repository/mysql"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var sf singleflight.Group

func GetCourseList(ctx context.Context) ([]model.Course, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	return mysqlrepo.GetCourseList(timeoutCtx)
}

func GetCourseById(ctx context.Context, id uint) (*model.Course, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*2)
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
		v, sfErr, shared := sf.Do(coursekey, func() (interface{}, error) {
			course, dbErr := mysqlrepo.GetCourseById(timeoutCtx, id)
			if dbErr == nil {
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

		if sfErr != nil {
			return nil, sfErr
		}

		if shared {
			global.Logger.Debug("singleflight成功合并了并发请求", zap.Uint("id", id))
		}

		sfResult := v.(*model.Course)
		return sfResult, nil
	}
	global.Logger.Error("Redis出错", zap.Error(err))
	return nil, ErrSystemBusy
}
