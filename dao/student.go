package dao

import (
	"context"
	"errors"
	"go-course/global"
	"go-course/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 创建学生函数
func CreateStu(timeoutCtx context.Context, student *model.Student) error {
	err := global.DB.WithContext(timeoutCtx).Create(student).Error
	if err != nil {
		global.Logger.Error("数据库出错", zap.Error(err))
	}
	return err
}

// 通过学号查询学生
func GetBySid(timeoutCtx context.Context, sid string) (*model.Student, error) {
	var student model.Student
	err := global.DB.WithContext(timeoutCtx).Where("sid = ?", sid).First(&student).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			global.Logger.Error("数据库查询异常", zap.String("sid", sid), zap.Error(err))
		}
		return nil, err
	}
	return &student, nil
}
