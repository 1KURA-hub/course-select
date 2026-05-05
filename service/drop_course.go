package service

import (
	"context"
	"errors"
	"fmt"
	"go-course/global"
	"go-course/model"
	redisrepo "go-course/repository/redis"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func DropCourse(ctx context.Context, studentID, courseID uint) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := global.DB.WithContext(timeoutCtx).Transaction(func(tx *gorm.DB) error {
		var selection model.Selection
		err := tx.Where("student_id = ? AND course_id = ? AND status = ?",
			studentID, courseID, model.SelectionStatusSelected).
			First(&selection).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSelectionNotFound
		}
		if err != nil {
			return err
		}

		if err := tx.Model(&selection).Update("status", model.SelectionStatusDropped).Error; err != nil {
			return err
		}

		db := tx.Model(&model.Course{}).
			Where("id = ?", courseID).
			Update("stock", gorm.Expr("stock + ?", 1))
		if db.Error != nil {
			return db.Error
		}
		if db.RowsAffected == 0 {
			return ErrSelectionNotFound
		}
		return nil
	})
	if err != nil {
		return err
	}

	stockKey := fmt.Sprintf("course:stock:%d", courseID)
	if err := global.RDB.Incr(timeoutCtx, stockKey).Err(); err != nil {
		global.Logger.Warn("退课后恢复Redis库存失败",
			zap.Uint("studentID", studentID),
			zap.Uint("courseID", courseID),
			zap.Error(err))
	}

	if err := redisrepo.DeleteSelectionRequestStatus(timeoutCtx, studentID, courseID); err != nil {
		global.Logger.Warn("退课后删除请求状态失败",
			zap.Uint("studentID", studentID),
			zap.Uint("courseID", courseID),
			zap.Error(err))
	}

	return nil
}
