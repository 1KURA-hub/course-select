package service

import (
	"errors"
	"go-course/dao"
	"go-course/global"
	"go-course/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Register(student *model.Student) error {
	// 通过Sid即学号查找学生
	_, err := dao.GetBySid(student.Sid)
	if err == nil {
		global.Logger.Debug("学生已经存在了", zap.String("sid", student.Sid))
		return ErrUserExist
	}
	// 如果err不是gorm查询为空 向前端返回模糊的错误原因
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrSystemBusy
	}

	// 此处salt仅示例使用
	student.Password = student.Password + "salt"

	err = dao.CreateStu(student)
	if err != nil {
		global.Logger.Error("数据库创建学生失败", zap.String("sid", student.Sid), zap.Error(err))
		return ErrSystemBusy
	}
	return nil
}

func Login(Sid string, password string) (*model.Student, error) {
	// 模糊返回错误原因
	savedStu, err := dao.GetBySid(Sid)
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
