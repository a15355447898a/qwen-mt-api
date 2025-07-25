package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"qwenmtapi/internal/controller"
	"qwenmtapi/internal/middleware"
)

func main() {
	s := g.Server()
	translateCtrl := controller.NewTranslateController()
	authMiddleware := middleware.NewAuthMiddleware()
	
	s.BindHandler("/", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"message": "Multi-Format Translation API powered by QWen",
			"version": "3.0.0",
			"endpoints": g.Map{
				"deeplx":  "POST /translate (DeepLX compatible)",
				"deepl":   "POST /v2/translate (DeepL compatible)",
				"legacy":  "POST /api/translate (legacy format)",
				"health":  "GET /health",
			},
		})
	})
	
	s.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status": "ok",
			"service": "qwenmtapi",
		})
	})
	
	// DeepLX 兼容接口 (带认证)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(authMiddleware.Middleware)
		group.POST("/translate", translateCtrl.DeepLXTranslate)
	})
	
	s.Group("/v2", func(group *ghttp.RouterGroup) {
		group.Middleware(authMiddleware.Middleware)
		group.POST("/translate", translateCtrl.Translate)
	})
	
	// 保持向后兼容 (带认证)
	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(authMiddleware.Middleware)
		group.POST("/translate", translateCtrl.Translate)
	})
	
	s.SetPort(8080)
	s.Run()
}