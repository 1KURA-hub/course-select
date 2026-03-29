package model

type Student struct {
	ID uint `gorm:"primarykey"`

	// Sid学号 唯一索引
	Sid string `gorm:"type:varchar(20);not null;uniqueIndex"`

	Name string `gorm:"type:varchar(20);not null"`

	Password string `gorm:"type:varchar(100);not null"`
}
