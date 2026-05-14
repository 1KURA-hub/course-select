package initialize

import (
	"go-course/global"
	"go-course/model"

	"go.uber.org/zap"
)

func SeedDemoCourses() {
	var count int64
	if err := global.DB.Model(&model.Course{}).Count(&count).Error; err != nil {
		global.Logger.Warn("查询课程种子数据失败", zap.Error(err))
		return
	}
	if count > 0 {
		return
	}

	courses := []model.Course{
		{ID: 1, Name: "高并发系统设计", Stock: 120, TeacherID: 1001},
		{ID: 2, Name: "分布式缓存实战", Stock: 80, TeacherID: 1002},
		{ID: 3, Name: "Go 后端工程化", Stock: 96, TeacherID: 1003},
		{ID: 4, Name: "Redis 与消息队列", Stock: 64, TeacherID: 1004},
		{ID: 5, Name: "MySQL 事务与索引优化", Stock: 72, TeacherID: 1005},
		{ID: 6, Name: "微服务架构设计", Stock: 48, TeacherID: 1006},
		{ID: 7, Name: "云原生应用开发", Stock: 90, TeacherID: 1007},
		{ID: 8, Name: "操作系统原理", Stock: 110, TeacherID: 1008},
		{ID: 9, Name: "数据结构与算法", Stock: 150, TeacherID: 1009},
		{ID: 10, Name: "计算机网络", Stock: 100, TeacherID: 1010},
	}
	if err := global.DB.Create(&courses).Error; err != nil {
		global.Logger.Warn("写入课程种子数据失败", zap.Error(err))
		return
	}
	global.Logger.Info("课程种子数据写入完成", zap.Int("count", len(courses)))
}
