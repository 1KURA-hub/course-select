package redisrepo

import (
	"context"
	"go-course/global"
	"time"
)

const (
	resultKeySuccessTTL = 5 * time.Minute
	resultKeyFailedTTL  = 24 * time.Hour
)

func MarkSelectionSuccess(ctx context.Context, studentID, courseID uint) error {
	return global.RDB.Set(ctx, ResultKey(studentID, courseID), 1, resultKeySuccessTTL).Err()
}

func MarkSelectionFailed(ctx context.Context, studentID, courseID uint) error {
	return global.RDB.Set(ctx, ResultKey(studentID, courseID), -1, resultKeyFailedTTL).Err()
}
