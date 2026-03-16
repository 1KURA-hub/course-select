package middleware

import (
	"go-course/global"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		status := c.Writer.Status()

		// 1. 服务器错误 (500+)  级别Error
		if status >= 500 {
			cost := time.Since(start)
			global.Logger.Error(path,
				zap.Int("status", status),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
				zap.Duration("cost", cost),
			)
			return
		}

		// 提前判断config里面是不是Debug级别 如果是生产模式 级别为Error 没有这行判断会因为go语言的函数特性
		// 先计算参数值 创建了参数 然后再判断级别 最后又被回收
		if !global.Logger.Core().Enabled(zap.DebugLevel) {
			return
		}

		// 2. 客户端错误 (400-499) 级别Debug
		if status >= 400 {
			cost := time.Since(start)
			global.Logger.Debug(path,
				zap.Int("status", status),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.Duration("cost", cost),
			)
			return
		}

		// 3. 正常请求 (200) 级别Debug
		cost := time.Since(start)
		global.Logger.Debug(path,
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Duration("cost", cost),
		)
	}
}
