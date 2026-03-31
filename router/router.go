package router

import (
	"go-course/api"
	"go-course/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	//r.Use(middleware.GinLogger())
	//pprof.Register(r)

	// CORS 跨域资源共享 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/register", api.Register)
	r.POST("/login", api.Login)
	r.GET("/courses", api.GetCourseList)
	r.GET("/courses/:id", middleware.Bloomfilter(), api.GetCourseById)

	// 路由组auth
	auth := r.Group("/auth")
	auth.Use(middleware.AuthMiddleware(), middleware.Bloomfilter())
	{
		auth.POST("/select/:id", api.SelectCourse)
		auth.GET("/result/:id", api.SelectResult)
	}

	return r
}
