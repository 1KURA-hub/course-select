package api

import (
	"errors"
	"go-course/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 获取课程列表接口
func GetCourseList(c *gin.Context) {
	courses, err := service.GetCourseList()
	if err != nil {
		// c.Error把具体错误存入Context ginlogger把err写入日志
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "系统繁忙 请稍后重试",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": courses,
	})

}

func GetCourseById(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "id格式错误",
		})
		return
	}
	course, err := service.GetCourseById(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"code": http.StatusNotFound,
				"msg":  "未找到对应课程",
			})
			return
		}
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "系统繁忙 请稍后重试",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": course,
	})

}
