package api

// 注册请求
type RegisterRequest struct {
	Sid      string `json:"sid"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// 登录请求
type LoginRequest struct {
	Sid      string `json:"sid"`
	Password string `json:"password"`
}

// 目标课程
type targetCourse struct {
	CourseID uint `json:"course_id"`
}
