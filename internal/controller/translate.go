package controller

import (
	"time"
	
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"qwenmtapi/internal/model"
	"qwenmtapi/internal/service"
)

type TranslateController struct {
	qwenService *service.QwenService
}

func NewTranslateController() *TranslateController {
	return &TranslateController{
		qwenService: service.NewQwenService(),
	}
}

func (c *TranslateController) Translate(r *ghttp.Request) {
	var req model.TranslateRequest
	if err := r.Parse(&req); err != nil {
		r.Response.Status = 400
		r.Response.WriteJsonExit(g.Map{
			"error": "Bad request: " + err.Error(),
		})
	}

	if err := g.Validator().Data(req).Run(r.Context()); err != nil {
		r.Response.Status = 400
		r.Response.WriteJsonExit(g.Map{
			"error": "Parameter validation failed: " + err.String(),
		})
	}

	result, err := c.qwenService.Translate(&req)
	if err != nil {
		r.Response.Status = 500
		r.Response.WriteJsonExit(g.Map{
			"error": "Translation failed: " + err.Error(),
		})
	}

	r.Response.WriteJsonExit(result)
}

func (c *TranslateController) DeepLXTranslate(r *ghttp.Request) {
	var req model.DeepLXRequest
	if err := r.Parse(&req); err != nil {
		r.Response.Status = 400
		r.Response.WriteJsonExit(model.DeepLXResponse{
			Code:    400,
			ID:      time.Now().UnixMilli(),
			Data:    "",
			Message: "Bad Request: " + err.Error(),
		})
	}

	if err := g.Validator().Data(req).Run(r.Context()); err != nil {
		r.Response.Status = 400
		r.Response.WriteJsonExit(model.DeepLXResponse{
			Code:    400,
			ID:      time.Now().UnixMilli(),
			Data:    "",
			Message: "Parameter validation failed: " + err.String(),
		})
	}

	result, err := c.qwenService.DeepLXTranslate(&req)
	if err != nil {
		r.Response.Status = 500
		r.Response.WriteJsonExit(model.DeepLXResponse{
			Code:    500,
			ID:      time.Now().UnixMilli(),
			Data:    "",
			Message: "Translation failed: " + err.Error(),
		})
	}

	r.Response.WriteJsonExit(result)
}