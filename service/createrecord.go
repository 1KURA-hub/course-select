package service

import (
	"context"
	"errors"
	"go-course/global"
	"go-course/model"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CreateRecord(timeoutCtx context.Context, studentID, courseID uint) error {
	// 开启事务 global.DB用tx
	return global.DB.WithContext(timeoutCtx).Transaction(func(tx *gorm.DB) error {
		var course model.Course
		course.ID = courseID

		db := tx.Model(&model.Course{}).
			Where("id = ? AND stock > ?", courseID, 0).
			Update("stock", gorm.Expr("stock-1"))

		if db.Error != nil {
			global.Logger.Error(ErrStockEmpty.Error(), zap.Error(db.Error))
			return errors.New("MySQL课程库存更新失败")
		}

		// 影响行数为0 更新失败
		if db.RowsAffected == 0 {
			global.Logger.Debug(ErrStockEmpty.Error())
			return ErrStockEmpty
		}

		var selection = &model.Selection{
			StudentID: studentID,
			CourseID:  course.ID,
			Status:    model.SelectionStatusSelected,
		}

		err := tx.Create(selection).Error
		if err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				var existing model.Selection
				queryErr := tx.Where("student_id = ? AND course_id = ?", studentID, courseID).First(&existing).Error
				if queryErr != nil {
					return queryErr
				}
				if existing.Status == model.SelectionStatusDropped {
					return tx.Model(&existing).Update("status", model.SelectionStatusSelected).Error
				}
				global.Logger.Error("重复选课", zap.Error(err))
				return ErrRepeatSelection
			}

			global.Logger.Error("数据库新建选课记录出错",
				zap.Uint("studentID", studentID),
				zap.Uint("course_id", courseID),
				zap.Error(err))
			return err
		}
		return nil
	})

}
