package service

import (
	"context"
	"fmt"
	"go-course/global"
	"time"
)

const (
	messageKeySuccessTTL = time.Hour
	resultKeySuccessTTL  = 5 * time.Minute
	resultKeyFailedTTL   = 24 * time.Hour
)

func RequestKey(studentID, courseID uint) string {
	return fmt.Sprintf("request:%d:%d", studentID, courseID)
}

func ResultKey(studentID, courseID uint) string {
	return fmt.Sprintf("res:%d:%d", studentID, courseID)
}

func MessageKey(studentID, courseID uint) string {
	return fmt.Sprintf("msg:%d:%d", studentID, courseID)
}

func MarkSelectionSuccess(ctx context.Context, studentID, courseID uint) error {
	pipe := global.RDB.TxPipeline()
	pipe.Set(ctx, MessageKey(studentID, courseID), "success", messageKeySuccessTTL)
	pipe.Set(ctx, ResultKey(studentID, courseID), 1, resultKeySuccessTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func MarkSelectionFailed(ctx context.Context, studentID, courseID uint) error {
	pipe := global.RDB.TxPipeline()
	pipe.Set(ctx, MessageKey(studentID, courseID), "success", messageKeySuccessTTL)
	pipe.Set(ctx, ResultKey(studentID, courseID), -1, resultKeyFailedTTL)
	_, err := pipe.Exec(ctx)
	return err
}
