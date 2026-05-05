package service

import (
	"context"
	"errors"
	"go-course/model"
	mysqlrepo "go-course/repository/mysql"
	redisrepo "go-course/repository/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	SelectionResultSuccess = "success"
	SelectionResultFailed  = "failed"
	SelectionResultPending = "pending"
	SelectionResultDropped = "dropped"
)

func QuerySelectResult(ctx context.Context, studentID, courseID uint) (string, error) {
	value, err := redisrepo.GetSelectionRequestStatus(ctx, studentID, courseID)
	if err == nil {
		switch value {
		case redisrepo.RequestStatusSuccess:
			return SelectionResultSuccess, nil
		case redisrepo.RequestStatusFailed:
			return SelectionResultFailed, nil
		case redisrepo.RequestStatusPending:
			return SelectionResultPending, nil
		default:
			return SelectionResultPending, nil
		}
	}

	if !errors.Is(err, redis.Nil) {
		return "", err
	}

	selection, dbErr := mysqlrepo.GetSelectionBySIDAndCID(studentID, courseID)
	if dbErr == nil && selection != nil && selection.Status == model.SelectionStatusSelected {
		return SelectionResultSuccess, nil
	}
	if dbErr == nil && selection != nil && selection.Status == model.SelectionStatusDropped {
		return SelectionResultDropped, nil
	}

	if errors.Is(dbErr, gorm.ErrRecordNotFound) {
		return SelectionResultPending, nil
	}

	if dbErr != nil {
		return "", dbErr
	}

	return SelectionResultPending, nil
}
