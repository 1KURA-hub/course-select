package initialize

import (
	"fmt"
	"go-course/global"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 初始化数据库连接
func InitMySQL() {
	m := global.Settings.MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		m.User,
		m.Password,
		m.Host,
		m.Port,
		m.DBName,
		m.Config,
	)
	var err error
	global.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 为了性能关闭自动外键
	})
	if err != nil {
		global.Logger.Fatal("初始化MySQL失败", zap.Error(err))
	}

	global.Logger.Info("初始化MySQL成功")

}
