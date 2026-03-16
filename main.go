package main

import (
	"fmt"
	"go-course/global"
	"go-course/initialize"
	"go-course/model"
	"go-course/mq"
	"go-course/router"
)

func main() {
	initialize.InitConfig()
	initialize.InitLogger()
	initialize.InitMySQL()
	initialize.InitRedis()
	initialize.InitBloomFilter()
	global.DB.AutoMigrate(&model.Student{}, &model.Course{}, &model.Selection{})
	initialize.InitRabbitMQ()

	mq.Consumer()

	r := router.InitRouter()
	r.Run(fmt.Sprintf(":%d", global.Settings.Server.Port))
}
