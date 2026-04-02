package service

import (
	"context"
	"errors"
	"go-course/dao"
	"go-course/global"
	"go-course/model"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Register(ctx context.Context, student *model.Student) error {
	// 网关ctx级联取消 timeoutCtx超时控制 业务逻辑生命周期最多2s
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := dao.GetBySid(timeoutCtx, student.Sid)
	if err == nil {
		global.Logger.Debug("学生已经存在了", zap.String("sid", student.Sid))
		return ErrUserExist
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrSystemBusy
	}

	// 此处salt仅示例使用
	student.Password = student.Password + "salt"

	err = dao.CreateStu(timeoutCtx, student)
	if err != nil {
		global.Logger.Error("数据库创建学生失败", zap.String("sid", student.Sid), zap.Error(err))
		return ErrSystemBusy
	}
	return nil
}

func Login(ctx context.Context, Sid string, password string) (*model.Student, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	savedStu, err := dao.GetBySid(timeoutCtx, Sid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.Logger.Debug("账号不存在", zap.String("sid", Sid))
			return nil, ErrUserPasswordError
		}
		return nil, ErrSystemBusy
	}

	if savedStu.Password != password+"salt" {
		global.Logger.Debug("学生输入密码错误", zap.String("sid", Sid))
		return nil, ErrUserPasswordError
	}
	return savedStu, nil
}
