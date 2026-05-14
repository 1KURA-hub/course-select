package router

import (
	"go-course/api"
	"go-course/middleware"
	"net/http"
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
		AllowOrigins:     []string{"http://localhost:8080", "http://localhost:5173", "http://localhost:5174", "http://localhost:5175", "http://127.0.0.1:5173", "http://127.0.0.1:5174", "http://127.0.0.1:5175"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/register", api.Register)
	r.POST("/login", api.Login)
	r.POST("/demo-login", api.DemoLogin)
	r.POST("/benchmark/start", api.StartBenchmark)
	r.GET("/benchmark/status", api.GetBenchmarkStatus)
	r.GET("/courses", api.GetCourseList)
	r.GET("/courses/:id", middleware.Bloomfilter(), api.GetCourseById)
	r.Static("/assets", "./web/dist/assets")
	r.GET("/", func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "ok"})
	})

	// 路由组auth
	auth := r.Group("/auth")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/selections", api.ListSelections)
		auth.POST("/select/:id", middleware.Bloomfilter(), api.SelectCourse)
		auth.DELETE("/select/:id", middleware.Bloomfilter(), api.DropCourse)
		auth.GET("/result/:id", middleware.Bloomfilter(), api.SelectResult)
	}

	r.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})

	return r
}
