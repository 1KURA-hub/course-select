package middleware

import (
	"go-course/global"
	"go-course/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware JWT 中间件鉴权
// 拦截所有请求，校验Authorization头 将已登录用户信息注入Context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Context中的请求头
		authHeader := c.GetHeader("Authorization")

		// 如果为空说明用户未登录 没有token 直接拦截
		if authHeader == "" {
			global.Logger.Debug("请求未携带token", zap.String("ip", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "用户未登录",
			})
			// 用户未登录 阻止调用后续api
			c.Abort()
			return
		}
		// 格式应为：Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			global.Logger.Debug("token格式错误", zap.String("ip", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "token格式错误",
			})
			c.Abort()
			return
		}
		// 对格式正确的token部分解析 验证签名和token有效期 获取token结构体中Claims的信息
		myClaims, err := utils.ParseToken(parts[1])
		if err != nil {
			global.Logger.Debug("token无效或者已过期", zap.Error(err), zap.String("ip", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "token无效或者已过期",
			})
			c.Abort()
			return
		}
		// 成功获取后 把Claims里面的用户ID和用户名字写入Context中
		// 后续api可以通过c.Get("user_id") 直接获取当前用户信息
		c.Set("user_id", myClaims.UserID)
		c.Set("username", myClaims.Username)

		// 继续调用api
		c.Next()

	}
}
