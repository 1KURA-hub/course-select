package api

import (
	"context"
	"errors"
	"fmt"
	"go-course/global"
	"go-course/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 选课接口
func SelectCourse(c *gin.Context) {
	// 通过在Context里面查找是否有"user_id"判断用户是否登录
	// 如果没有 说明验证token的中间件没能把token中携带的claims信息写入Context
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": http.StatusUnauthorized,
			"msg":  "用户未登录",
		})
		return
	}
	// val是空接口 val类型断言
	studentID, ok := val.(uint)
	if !ok {
		c.Error(errors.New("SelectCourse: UserID类型断言失败"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "系统繁忙 请稍后重试",
		})
		return
	}
	info := targetCourse{}
	err := c.ShouldBindJSON(&info)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "课程参数错误",
		})
		return
	}

	err = service.SelectCourse(studentID, info.CourseID)
	if err != nil {
		if errors.Is(err, service.ErrSystemBusy) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  err.Error(),
		})
		return
	}
	// MQ消息发送成功后 前端返回 排队中
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "排队中",
	})
}

// 查询选课结果接口
func SelectResult(c *gin.Context) {
	// 验证token
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": http.StatusUnauthorized,
			"msg":  "用户未登录",
		})
		return
	}
	// 空接口类型断言为uint
	studentID, ok := val.(uint)
	if !ok {
		// 错误写入Context 最后写入日志
		c.Error(errors.New("SelectResult: UserID类型断言失败"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "系统错误",
		})
		return
	}

	// 在url的query获取courseID(string类型)
	courseIDstr := c.Query("course_id")
	if courseIDstr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "用户未填写course_id",
		})
		return
	}
	courseID, err := strconv.Atoi(courseIDstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "用户填入course_id格式错误",
		})
		return
	}
	// 协程超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 消费者接收消息完成数据库的操作后会在Redis里面生成key为res:studentID:courseID value为1的记录
	// 在Redis中查找res:1:1这样的key对应的value
	key := fmt.Sprintf("res:%d:%d", studentID, courseID)

	value, err := global.RDB.Get(ctx, key).Result()
	if err != nil {
		//Redis查询结果为空 说明消息还没发送过来或者对数据库的操作还在排队
		if err == redis.Nil {
			c.JSON(http.StatusOK, gin.H{
				"data": nil,
				"msg":  "排队中",
			})
			return
		}
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "系统出错",
		})
		return
	}
	//value不为空 说明Redis里面有记录 选课成功了
	if value == "1" {
		c.JSON(http.StatusOK, gin.H{
			"code":      http.StatusOK,
			"msg":       "抢课成功",
			"course_id": courseID,
		})
	}

	if value == "-1" {
		c.JSON(http.StatusOK, gin.H{
			"code":      http.StatusOK,
			"msg":       "抢课失败",
			"course_id": courseID,
		})
	}
}
