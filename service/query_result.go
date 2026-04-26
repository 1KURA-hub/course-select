package service

import (
	"context"
	"errors"
	"go-course/global"
	mysqlrepo "go-course/repository/mysql"
	redisrepo "go-course/repository/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	SelectionResultSuccess = "success"
	SelectionResultFailed  = "failed"
	SelectionResultPending = "pending"
)

func QuerySelectResult(ctx context.Context, studentID, courseID uint) (string, error) {
	value, err := global.RDB.Get(ctx, redisrepo.ResultKey(studentID, courseID)).Result()
	if err == nil {
		switch value {
		case "1":
			return SelectionResultSuccess, nil
		case "-1":
			return SelectionResultFailed, nil
		default:
			return SelectionResultPending, nil
		}
	}

	if !errors.Is(err, redis.Nil) {
		return "", err
	}

	selection, dbErr := mysqlrepo.GetSelectionBySIDAndCID(studentID, courseID)
	if dbErr == nil && selection != nil {
		return SelectionResultSuccess, nil
	}

	if errors.Is(dbErr, gorm.ErrRecordNotFound) {
		return SelectionResultPending, nil
	}

	if dbErr != nil {
		return "", dbErr
	}

	return SelectionResultPending, nil
}
