package api

import (
	"errors"
	"go-course/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
	// /select/:id url参数中获取课程id
	courseIDstr := c.Param("id")
	courseID, err := strconv.Atoi(courseIDstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "课程参数错误",
		})
		return
	}

	err = service.SelectCourse(c.Request.Context(), studentID, uint(courseID))
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

	courseIDstr := c.Param("id")
	courseID, err := strconv.Atoi(courseIDstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "课程参数错误",
		})
		return
	}

	result, err := service.QuerySelectResult(c.Request.Context(), studentID, uint(courseID))
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "系统出错",
		})
		return
	}

	if result == service.SelectionResultSuccess {
		c.JSON(http.StatusOK, gin.H{
			"code":      http.StatusOK,
			"msg":       "抢课成功",
			"course_id": courseID,
		})
		return
	}

	if result == service.SelectionResultFailed {
		c.JSON(http.StatusOK, gin.H{
			"code":      http.StatusOK,
			"msg":       "抢课失败",
			"course_id": courseID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": nil,
		"msg":  "排队中",
	})
}
