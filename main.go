package main

import (
	"fmt"
	"go-course/global"
	"go-course/initialize"
	"go-course/model"
	"go-course/mq"
	"go-course/router"
	"net/http"
	_ "net/http/pprof"

	"go.uber.org/zap"
)

func main() {
	initialize.InitConfig()
	initialize.InitLogger()
	initialize.InitMySQL()

	global.DB.AutoMigrate(&model.Student{}, &model.Course{}, &model.Selection{})

	initialize.InitRedis()
	initialize.InitBloomFilter()
	initialize.InitRabbitMQ()
	go func() {
		global.Logger.Info("pprof 内网监控启动于 :6060 端口")
		// 如果 ListenAndServe 报错退出 打印日志
		if err := http.ListenAndServe(":6060", nil); err != nil {
			global.Logger.Error("pprof 监控服务异常退出", zap.Error(err))
		}
	}()

	mq.StartRelay()
	mq.Consumer()

	r := router.InitRouter()
	r.Run(fmt.Sprintf(":%d", global.Settings.Server.Port))
}
