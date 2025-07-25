package model

type TranslateRequest struct {
	Text       []string `json:"text" v:"required|min-length:1#请输入要翻译的文本|文本不能为空"`
	SourceLang string   `json:"source_lang"`
	TargetLang string   `json:"target_lang" v:"required#请选择目标语言"`
	Context    string   `json:"context"`
}

type Translation struct {
	DetectedSourceLanguage string `json:"detected_source_language"`
	Text                   string `json:"text"`
}

type TranslateResponse struct {
	Translations []Translation `json:"translations"`
}

type DeepLXRequest struct {
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang" v:"required#请选择目标语言"`
	Text       string `json:"text" v:"required|min-length:1#请输入要翻译的文本|文本不能为空"`
}

type DeepLXResponse struct {
	Code    int    `json:"code"`
	ID      int64  `json:"id"`
	Data    string `json:"data"`
	Message string `json:"message,omitempty"`
}

type QwenJoinRequest struct {
	Data        []string    `json:"data"`
	EventData   interface{} `json:"event_data"`
	FnIndex     int         `json:"fn_index"`
	TriggerID   int         `json:"trigger_id"`
	DataType    []string    `json:"dataType"`
	SessionHash string      `json:"session_hash"`
}

type QwenJoinResponse struct {
	EventID     string `json:"event_id"`
	Rank        int    `json:"rank"`
	QueueFull   bool   `json:"queue_full"`
	Success     bool   `json:"success"`
	SessionHash string `json:"session_hash"`
}