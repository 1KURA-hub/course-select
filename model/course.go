package model

import "gorm.io/gorm"

type Course struct {
	gorm.Model

	Name string `gorm:"type:varchar(50);not null"`

	TotalStock int `gorm:"not null"`

	TeacherID int `gorm:"index"`
}
