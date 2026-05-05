package model

const (
	SelectionStatusSelected = 1
	SelectionStatusDropped  = 2
)

type Selection struct {
	ID uint `gorm:"primarykey"`
	// 选课表 student_id和course_id建立唯一联合索引 保证不重复选课
	StudentID uint `gorm:"index;uniqueIndex:idx_student_course;comment:学生ID"`

	CourseID uint `gorm:"index;uniqueIndex:idx_student_course;comment:课程ID"`

	Status int `gorm:"type:tinyint;default:1;comment:选课状态"`
}
