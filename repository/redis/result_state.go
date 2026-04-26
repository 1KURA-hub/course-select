package redisrepo

import (
	"context"
	"go-course/global"
	"time"
)

const (
	messageKeySuccessTTL = time.Hour
	resultKeySuccessTTL  = 5 * time.Minute
	resultKeyFailedTTL   = 24 * time.Hour
)

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
