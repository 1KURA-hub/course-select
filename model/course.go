package model

type Course struct {
	ID uint `gorm:"primarykey"`

	Name string `gorm:"type:varchar(50);not null"`

	Stock int `gorm:"not null"`

	TeacherID int `gorm:"index"`
}
