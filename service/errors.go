package service

import "errors"

var (
	// 用户异常
	ErrUserExist         = errors.New("该学生已注册")
	ErrUserPasswordError = errors.New("账号或密码错误")

	// 选课异常
	ErrRepeatRequest     = errors.New("重复请求")
	ErrStockEmpty        = errors.New("课程库存不足")
	ErrRepeatSelection   = errors.New("不可重复选课")
	ErrSelectionNotFound = errors.New("未找到可退课记录")

	// 系统异常
	ErrSystemBusy = errors.New("系统繁忙 请稍后再试")
)
