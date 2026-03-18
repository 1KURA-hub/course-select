package service

import (
	"context"
	"errors"
	"go-course/global"
	"go-course/model"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 创建选课记录
func CreateRecord(timeoutCtx context.Context, studentID, courseID uint) error {
	// 开启事务 global.DB用tx
	return global.DB.WithContext(timeoutCtx).Transaction(func(tx *gorm.DB) error {
		var course model.Course
		course.ID = courseID

		// select for update 当前读 对库存上记录锁
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", course.ID).First(&course).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				global.Logger.Debug("课程不存在", zap.Uint("course_id", courseID))
				return errors.New("课程不存在")
			}
			global.Logger.Error("数据库查询课程时出错", zap.Uint("course_id", courseID), zap.Error(err))
			return ErrSystemBusy
		}

		if course.TotalStock > 0 {
			// 更新课程库存
			course.TotalStock--
			err = tx.Model(&course).Update("total_stock", course.TotalStock).Error
			if err != nil {
				global.Logger.Error("数据库更新课程库存时出错", zap.Uint("course_id", courseID), zap.Error(err))
				return ErrSystemBusy
			}

			// 创建选课记录实例
			var selection = &model.Selection{
				StudentID: studentID,
				CourseID:  course.ID,
			}

			// 选课记录写入数据库
			err = tx.Create(selection).Error
			if err != nil {
				// 判断是否是唯一索引冲突
				if strings.Contains(err.Error(), "Duplicate entry") {
					global.Logger.Error("重复选课", zap.Error(err))
					return ErrRepeatSelection
				}

				global.Logger.Error("数据库新建选课记录出错",
					zap.Uint("studentID", studentID),
					zap.Uint("course_id", courseID),
					zap.Error(err))
				return err
			}
			return nil
		}
		global.Logger.Debug("课程库存不足", zap.Uint("course_id", courseID))
		return ErrStockEmpty
	})

}
