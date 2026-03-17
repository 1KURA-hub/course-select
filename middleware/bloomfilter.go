package middleware

import (
	"fmt"
	"go-course/global"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Bloomfilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 布隆过滤器 快速过滤不存在的课程ID
		if !global.CourseBloomFilter.TestString(fmt.Sprintf("%d", c.Param("id"))) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "不存在的课程",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
