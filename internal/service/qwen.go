package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/guid"

	"qwenmtapi/internal/model"
)

type QwenService struct {
	client *gclient.Client
	baseURL string
	langMap map[string]string
	rateLimiter chan struct{} // 限流通道
}

func NewQwenService() *QwenService {
	client := g.Client()
	client.SetTimeout(60 * time.Second) // 增加超时时间到60秒
	
	// 创建语言映射表 - 仅支持简体中文和英语
	langMap := map[string]string{
		// 中文名称到标准代码映射
		"自动检测":  "auto",
		"自动":    "auto", 
		"简体中文": "ZH",
		"英语":   "EN",
		"英文":   "EN",
		
		// 标准代码保持不变
		"auto": "auto",
		"ZH": "ZH",
		"EN": "EN",
	}
	
	// 创建限流通道，最多允许2个并发请求（减少并发以避免会话冲突）
	rateLimiter := make(chan struct{}, 2)
	
	return &QwenService{
		client:      client,
		baseURL:     "https://qwen-qwen3-mt-demo.ms.show",
		langMap:     langMap,
		rateLimiter: rateLimiter,
	}
}

func (s *QwenService) Translate(req *model.TranslateRequest) (*model.TranslateResponse, error) {
	ctx := gctx.New()
	var translations []model.Translation
	
	// 映射语言代码
	sourceLang := s.mapLanguage(req.SourceLang)
	targetLang := s.mapLanguage(req.TargetLang)
	
	for _, text := range req.Text {
		result, detectedLang, err := s.translateSingleText(ctx, text, sourceLang, targetLang)
		if err != nil {
			return nil, err
		}
		
		translations = append(translations, model.Translation{
			DetectedSourceLanguage: detectedLang,
			Text:                   result,
		})
	}
	
	return &model.TranslateResponse{
		Translations: translations,
	}, nil
}

// mapLanguage 将中文语言名称映射为千问API需要的格式
func (s *QwenService) mapLanguage(lang string) string {
	// 千问API使用的语言映射
	qwenLangMap := map[string]string{
		"auto": "自动检测",
		"ZH":   "简体中文", 
		"EN":   "英语",
	}
	
	// 先通过标准映射获取代码
	if mappedLang, exists := s.langMap[lang]; exists {
		// 再映射为千问API格式
		if qwenLang, exists := qwenLangMap[mappedLang]; exists {
			return qwenLang
		}
		return mappedLang
	}
	
	// 如果找不到映射，返回原值
	return lang
}

func (s *QwenService) translateSingleText(ctx context.Context, text, sourceLang, targetLang string) (string, string, error) {
	// 限流控制
	s.rateLimiter <- struct{}{}
	defer func() { <-s.rateLimiter }()
	
	// 重试机制：最多重试3次
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			// 重试前等待一段时间，避免立即重试
			time.Sleep(time.Duration(attempt) * time.Second)
			g.Log().Info(ctx, fmt.Sprintf("Retrying translation attempt %d", attempt+1))
		}
		
		result, detectedLang, err := s.attemptTranslation(ctx, text, sourceLang, targetLang)
		if err == nil {
			return result, detectedLang, nil
		}
		
		lastErr = err
		// 如果是会话错误，继续重试；其他错误直接返回
		if !strings.Contains(err.Error(), "Session not found") && 
		   !strings.Contains(err.Error(), "unexpected_error") {
			break
		}
	}
	
	return "", "", fmt.Errorf("translation failed after 3 attempts: %v", lastErr)
}

func (s *QwenService) attemptTranslation(ctx context.Context, text, sourceLang, targetLang string) (string, string, error) {
	sessionHash := strings.ReplaceAll(guid.S(), "-", "")[:12]
	
	// Step 1: Join queue
	joinReq := &model.QwenJoinRequest{
		Data:        []string{text, sourceLang, targetLang},
		EventData:   nil,
		FnIndex:     2,
		TriggerID:   11,
		DataType:    []string{"textbox", "dropdown", "dropdown"},
		SessionHash: sessionHash,
	}
	
	joinURL := fmt.Sprintf("%s/gradio_api/queue/join?t=%d&__theme=light&backend_url=%%2F", 
		s.baseURL, time.Now().UnixMilli())
	
	headers := map[string]string{
		"Accept":                    "*/*",
		"Accept-Language":           "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
		"Content-Type":             "application/json",
		"Origin":                   s.baseURL,
		"Priority":                 "u=1, i",
		"Referer":                  fmt.Sprintf("%s/?t=%d&__theme=light&backend_url=/", s.baseURL, time.Now().UnixMilli()),
		"Sec-Ch-Ua":                `"Not)A;Brand";v="8", "Chromium";v="138", "Microsoft Edge";v="138"`,
		"Sec-Ch-Ua-Mobile":         "?0",
		"Sec-Ch-Ua-Platform":       `"macOS"`,
		"Sec-Fetch-Dest":           "empty",
		"Sec-Fetch-Mode":           "cors",
		"Sec-Fetch-Site":           "same-origin",
		"Sec-Fetch-Storage-Access": "active",
		"User-Agent":               "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36 Edg/138.0.0.0",
		"X-Studio-Token":           "",
	}
	
	joinResp, err := s.client.Header(headers).Post(ctx, joinURL, joinReq)
	if err != nil {
		return "", "", fmt.Errorf("failed to join queue: %v", err)
	}
	defer joinResp.Close()
	
	respBody := joinResp.ReadAll()
	g.Log().Info(ctx, "Join response:", string(respBody))
	
	var joinResult model.QwenJoinResponse
	if err := json.Unmarshal(respBody, &joinResult); err != nil {
		return "", "", fmt.Errorf("failed to parse join response: %v, body: %s", err, string(respBody))
	}
	
	// 检查是否有event_id，有的话说明加入队列成功
	if joinResult.EventID == "" {
		return "", "", fmt.Errorf("failed to join queue, response: %s", string(respBody))
	}
	
	// Step 2: Get translation result
	dataURL := fmt.Sprintf("%s/gradio_api/queue/data?session_hash=%s&studio_token=", 
		s.baseURL, sessionHash)
	
	dataHeaders := map[string]string{
		"Accept":                   "text/event-stream",
		"Accept-Language":          "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
		"Content-Type":             "application/json",
		"Priority":                 "u=1, i",
		"Referer":                  fmt.Sprintf("%s/?t=%d&__theme=light&backend_url=/", s.baseURL, time.Now().UnixMilli()),
		"Sec-Ch-Ua":                `"Not)A;Brand";v="8", "Chromium";v="138", "Microsoft Edge";v="138"`,
		"Sec-Ch-Ua-Mobile":         "?0",
		"Sec-Ch-Ua-Platform":       `"macOS"`,
		"Sec-Fetch-Dest":           "empty",
		"Sec-Fetch-Mode":           "cors",
		"Sec-Fetch-Site":           "same-origin",
		"Sec-Fetch-Storage-Access": "active",
		"User-Agent":               "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36 Edg/138.0.0.0",
		"X-Studio-Token":           "",
	}
	
	dataResp, err := s.client.Header(dataHeaders).Get(ctx, dataURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to get translation data: %v", err)
	}
	defer dataResp.Close()
	
	result, err := s.parseEventStream(dataResp)
	if err != nil {
		return "", "", err
	}
	
	// 返回翻译结果和检测到的源语言（如果没有指定源语言，则使用目标语言作为占位符）
	detectedLang := sourceLang
	if detectedLang == "" {
		detectedLang = "auto"
	}
	
	return result, detectedLang, nil
}

func (s *QwenService) DeepLXTranslate(req *model.DeepLXRequest) (*model.DeepLXResponse, error) {
	ctx := gctx.New()
	requestID := time.Now().UnixMilli()
	
	// 映射语言代码
	sourceLang := s.mapLanguage(req.SourceLang)
	targetLang := s.mapLanguage(req.TargetLang)
	
	// 获取主要翻译结果
	result, _, err := s.translateSingleText(ctx, req.Text, sourceLang, targetLang)
	if err != nil {
		return nil, err
	}
	
	return &model.DeepLXResponse{
		Code: 200,
		ID:   requestID,
		Data: result,
	}, nil
}



func (s *QwenService) parseEventStream(resp *gclient.Response) (string, error) {
	scanner := bufio.NewScanner(resp.Body)
	ctx := gctx.New()
	
	for scanner.Scan() {
		line := scanner.Text()
		g.Log().Debug(ctx, "Event stream line:", line)
		
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}
			
			var eventData map[string]interface{}
			if err := json.Unmarshal([]byte(data), &eventData); err != nil {
				g.Log().Debug(ctx, "Failed to parse event data:", data, "error:", err)
				continue
			}
			
			g.Log().Debug(ctx, "Parsed event data:", eventData)
			
			// 检查是否有错误消息
			if msg, ok := eventData["msg"].(string); ok {
				if msg == "unexpected_error" {
					if message, ok := eventData["message"].(string); ok {
						return "", fmt.Errorf("unexpected_error: %s", message)
					}
					return "", fmt.Errorf("unexpected_error occurred")
				}
				
				if msg == "process_completed" {
					if output, ok := eventData["output"].(map[string]interface{}); ok {
						// 检查输出中是否有错误
						if errMsg, ok := output["error"].(string); ok {
							return "", fmt.Errorf("process_error: %s", errMsg)
						}
						
						if dataArray, ok := output["data"].([]interface{}); ok && len(dataArray) > 0 {
							if result, ok := dataArray[0].(string); ok {
								return result, nil
							}
						}
					}
				}
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading event stream: %v", err)
	}
	
	return "", fmt.Errorf("no translation result found")
}