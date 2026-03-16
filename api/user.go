package api

import (
	"errors"
	"go-course/model"
	"go-course/service"
	"go-course/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 注册接口
func Register(c *gin.Context) {
	var request RegisterRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "用户注册时填写格式错误",
		})
		return
	}
	// 把request结构体从Context里面获取的字段写入student结构体
	student := &model.Student{
		Sid:      request.Sid,
		Password: request.Password,
		Name:     request.Name,
	}
	err = service.Register(student)
	if err != nil {
		if err == service.ErrUserExist {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "该学生已注册",
			})
			return
		}
		// 把具体错误原因写入Context ginlogger会获取错误写入日志
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
		return

	}
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "注册成功",
	})
}

// 登录接口
func Login(c *gin.Context) {
	var request LoginRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "用户登录时填写格式错误",
		})
		return
	}
	// 模糊返回错误原因
	savedStu, err := service.Login(request.Sid, request.Password)
	if errors.Is(err, service.ErrUserPasswordError) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  err.Error(),
		})
		return
	}
	// 其他情况就是数据库挂了
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
		return
	}
	// 登录成功生成token 包含学生ID和姓名信息
	var tokenstr string
	tokenstr, err = utils.GenToken(savedStu.ID, savedStu.Name)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "生成token时系统出错",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  http.StatusOK,
		"token": tokenstr,
		"name":  savedStu.Name,
		"id":    savedStu.ID,
	})

}
