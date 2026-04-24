package service

import (
	"context"
	"go-course/global"
	"go-course/model"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CompensateRecord(timeoutCtx context.Context, studentID, courseID uint) error {
	return global.DB.WithContext(timeoutCtx).Transaction(func(tx *gorm.DB) error {
		selection := &model.Selection{
			StudentID: studentID,
			CourseID:  courseID,
		}

		err := tx.Create(selection).Error
		if err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				global.Logger.Warn("死信补偿命中唯一索引，说明记录已存在",
					zap.Uint("studentID", studentID),
					zap.Uint("courseID", courseID),
					zap.Error(err))
				return ErrRepeatSelection
			}

			global.Logger.Error("死信补偿写入选课记录失败",
				zap.Uint("studentID", studentID),
				zap.Uint("courseID", courseID),
				zap.Error(err))
			return err
		}
		return nil
	})
}
