package dao

import (
	"context"
	"errors"
	"go-course/global"
	"go-course/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 获取课程列表函数
func GetCourseList(timeoutCtx context.Context) ([]model.Course, error) {
	var courses []model.Course
	// gorm会自动通过&courses的类型Course 在数据库中找course的表 存入courses课程切片
	err := global.DB.WithContext(timeoutCtx).Find(&courses).Error
	if err != nil {
		global.Logger.Error("数据库出错", zap.Error(err))
		return nil, err
	}
	return courses, nil
}

func GetCourseById(timeoutCtx context.Context, id uint) (*model.Course, error) {
	var course model.Course
	err := global.DB.WithContext(timeoutCtx).First(&course, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		global.Logger.Error("数据库出错", zap.Error(err))
		return nil, err
	}
	return &course, nil
}
