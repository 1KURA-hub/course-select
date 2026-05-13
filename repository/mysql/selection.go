package mysqlrepo

import (
	"context"
	"errors"
	"go-course/global"
	"go-course/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 通过学生ID和课程ID获取选课记录 主要是判断学生是否重复选课
func GetSelectionBySIDAndCID(SID, CID uint) (*model.Selection, error) {
	var selection model.Selection
	err := global.DB.Where("student_id = ? and course_id = ?", SID, CID).First(&selection).Error
	if err != nil {
		// 首次选课
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		global.Logger.Error("数据库出错", zap.Error(err))
		return nil, err
	}
	return &selection, nil
}

type StudentSelection struct {
	SelectionID uint   `json:"selection_id"`
	StudentID   uint   `json:"student_id"`
	CourseID    uint   `json:"course_id"`
	Status      int    `json:"status"`
	StatusText  string `json:"status_text" gorm:"-"`
	CourseName  string `json:"course_name"`
	TeacherID   int    `json:"teacher_id"`
}

func ListSelectionsByStudentID(ctx context.Context, studentID uint) ([]StudentSelection, error) {
	var selections []StudentSelection
	err := global.DB.WithContext(ctx).
		Table("selections").
		Select("selections.id AS selection_id, selections.student_id, selections.course_id, selections.status, courses.name AS course_name, courses.teacher_id").
		Joins("LEFT JOIN courses ON courses.id = selections.course_id").
		Where("selections.student_id = ? AND selections.status <> ?", studentID, model.SelectionStatusDropped).
		Order("selections.id DESC").
		Scan(&selections).Error
	if err != nil {
		global.Logger.Error("查询学生选课记录失败", zap.Uint("studentID", studentID), zap.Error(err))
		return nil, err
	}
	for i := range selections {
		selections[i].StatusText = SelectionStatusText(selections[i].Status)
	}
	return selections, nil
}

func SelectionStatusText(status int) string {
	switch status {
	case model.SelectionStatusSelected:
		return "选课成功"
	case model.SelectionStatusDropped:
		return "已退课"
	default:
		return "未知状态"
	}
}
