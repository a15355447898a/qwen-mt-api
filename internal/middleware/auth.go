package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"qwenmtapi/internal/model"
)

type AuthMiddleware struct {
	enabled    bool
	validKeys  map[string]bool
	required   bool
}

func NewAuthMiddleware() *AuthMiddleware {
	auth := &AuthMiddleware{
		enabled:   false,
		validKeys: make(map[string]bool),
		required:  false,
	}

	// 检查是否启用认证
	if os.Getenv("AUTH_ENABLED") == "true" {
		auth.enabled = true
		auth.required = true
	}

	// 加载API Keys
	auth.loadAPIKeys()

	return auth
}

func (auth *AuthMiddleware) loadAPIKeys() {
	// 从环境变量加载单个key
	if apiKey := os.Getenv("API_KEY"); apiKey != "" {
		auth.validKeys[apiKey] = true
	}

	// 从环境变量加载多个keys (逗号分隔)
	if apiKeys := os.Getenv("API_KEYS"); apiKeys != "" {
		keys := strings.Split(apiKeys, ",")
		for _, key := range keys {
			key = strings.TrimSpace(key)
			if key != "" {
				auth.validKeys[key] = true
			}
		}
	}

	// 如果没有配置key但启用了认证，添加默认key
	if auth.enabled && len(auth.validKeys) == 0 {
		auth.validKeys["sk-default"] = true
		g.Log().Warning(nil, "No API keys configured, using default key: sk-default")
	}
}

func (auth *AuthMiddleware) Middleware(r *ghttp.Request) {
	// 如果未启用认证，直接通过
	if !auth.enabled {
		r.Middleware.Next()
		return
	}

	// 提取API Key
	apiKey := auth.extractAPIKey(r)

	// 验证Key
	if !auth.validateKey(apiKey) {
		auth.unauthorized(r)
		return
	}

	// 认证成功，继续处理
	r.Middleware.Next()
}

func (auth *AuthMiddleware) extractAPIKey(r *ghttp.Request) string {
	// 1. Authorization: DeepL-Auth-Key [key]
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		if strings.HasPrefix(authHeader, "DeepL-Auth-Key ") {
			return strings.TrimPrefix(authHeader, "DeepL-Auth-Key ")
		}
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// 2. X-API-Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// 3. api_key query parameter
	if apiKey := r.Get("api_key").String(); apiKey != "" {
		return apiKey
	}

	return ""
}

func (auth *AuthMiddleware) validateKey(key string) bool {
	if key == "" {
		return false
	}
	return auth.validKeys[key]
}

func (auth *AuthMiddleware) unauthorized(r *ghttp.Request) {
	r.Response.Status = 401

	// 根据请求路径返回不同格式的错误
	path := r.URL.Path

	if strings.Contains(path, "/v2/translate") {
		// DeepL 格式错误
		r.Response.WriteJsonExit(g.Map{
			"error": "Authorization failed. Please supply a valid DeepL-Auth-Key via the Authorization header.",
		})
	} else if strings.Contains(path, "/translate") && !strings.Contains(path, "/api/") {
		// DeepLX 格式错误
		r.Response.WriteJsonExit(model.DeepLXResponse{
			Code:    401,
			ID:      time.Now().UnixMilli(),
			Data:    "",
			Message: "Unauthorized: Invalid or missing API key",
		})
	} else {
		// 原始格式错误
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "Unauthorized: Missing or invalid API key",
			"data":    "",
		})
	}
}