package service

import (
	"context"
	mysqlrepo "go-course/repository/mysql"
	"time"
)

func ListStudentSelections(ctx context.Context, studentID uint) ([]mysqlrepo.StudentSelection, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return mysqlrepo.ListSelectionsByStudentID(timeoutCtx, studentID)
}
