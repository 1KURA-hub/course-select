package redisrepo

import (
	"context"
	"go-course/global"
	"time"
)

const (
	requestStatusTTL = 24 * time.Hour
)

const (
	RequestStatusPending = "pending"
	RequestStatusSuccess = "success"
	RequestStatusFailed  = "failed"
)

func GetSelectionRequestStatus(ctx context.Context, studentID, courseID uint) (string, error) {
	return global.RDB.Get(ctx, RequestKey(studentID, courseID)).Result()
}

func MarkSelectionRequestPending(ctx context.Context, studentID, courseID uint) error {
	return global.RDB.Set(ctx, RequestKey(studentID, courseID), RequestStatusPending, requestStatusTTL).Err()
}

func MarkSelectionRequestSuccess(ctx context.Context, studentID, courseID uint) error {
	return global.RDB.Set(ctx, RequestKey(studentID, courseID), RequestStatusSuccess, requestStatusTTL).Err()
}

func MarkSelectionRequestFailed(ctx context.Context, studentID, courseID uint) error {
	return global.RDB.Set(ctx, RequestKey(studentID, courseID), RequestStatusFailed, requestStatusTTL).Err()
}
