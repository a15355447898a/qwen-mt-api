package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"
)

type CORSMiddleware struct{}

func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{}
}

func (cors *CORSMiddleware) Middleware(r *ghttp.Request) {
	// 设置CORS头
	r.Response.CORSDefault()
	
	// 如果是预检请求，直接返回
	if r.Method == "OPTIONS" {
		r.Response.Status = 200
		return
	}
	
	// 继续处理请求
	r.Middleware.Next()
}