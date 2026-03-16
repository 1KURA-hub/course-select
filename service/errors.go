package service

import "errors"

// 定义部分全局业务逻辑error
var (
	// 用户异常
	ErrUserExist         = errors.New("该学生已注册")
	ErrUserPasswordError = errors.New("账号或密码错误")

	// 选课异常
	ErrRepeatRequest   = errors.New("重复请求")
	ErrStockEmpty      = errors.New("课程库存不足")
	ErrRepeatSelection = errors.New("不可重复选课")

	// 系统异常
	ErrSystemBusy = errors.New("系统繁忙 请稍后再试")
)
