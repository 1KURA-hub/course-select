package service

import (
	"context"
	"go-course/global"

	"go.uber.org/zap"
)

func CompensateRecord(timeoutCtx context.Context, studentID, courseID uint) error {
	err := CreateRecord(timeoutCtx, studentID, courseID)
	if err != nil {
		if err == ErrRepeatSelection {
			global.Logger.Warn("死信补偿命中唯一索引，说明记录已存在",
				zap.Uint("studentID", studentID),
				zap.Uint("courseID", courseID),
				zap.Error(err))
			return err
		}

		global.Logger.Error("死信补偿重放完整落库事务失败",
			zap.Uint("studentID", studentID),
			zap.Uint("courseID", courseID),
			zap.Error(err))
		return err
	}

	global.Logger.Info("死信补偿已重放完整落库事务",
		zap.Uint("studentID", studentID),
		zap.Uint("courseID", courseID))
	return nil
}
