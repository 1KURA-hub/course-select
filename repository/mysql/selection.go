package mysqlrepo

import (
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
