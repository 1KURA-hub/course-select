package redisrepo

import "fmt"

const SelectStreamKey = "select:stream"

func RequestKey(studentID, courseID uint) string {
	return fmt.Sprintf("request:%d:%d", studentID, courseID)
}
