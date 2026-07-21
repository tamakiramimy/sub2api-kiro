package kiro

type Model struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
}

var DefaultModels = []Model{
	{ID: "claude-opus-4-8", Type: "model", DisplayName: "Claude Opus 4.8"},
	{ID: "claude-opus-4-8-thinking", Type: "model", DisplayName: "Claude Opus 4.8 (Thinking)"},
	{ID: "claude-opus-4-7", Type: "model", DisplayName: "Claude Opus 4.7"},
	{ID: "claude-opus-4-7-thinking", Type: "model", DisplayName: "Claude Opus 4.7 (Thinking)"},
	{ID: "claude-sonnet-5-0", Type: "model", DisplayName: "Claude Sonnet 5.0"},
	{ID: "claude-sonnet-5-0-thinking", Type: "model", DisplayName: "Claude Sonnet 5.0 (Thinking)"},
	{ID: "claude-opus-4-6", Type: "model", DisplayName: "Claude Opus 4.6"},
	{ID: "claude-opus-4-6-thinking", Type: "model", DisplayName: "Claude Opus 4.6 (Thinking)"},
	{ID: "claude-sonnet-4-6", Type: "model", DisplayName: "Claude Sonnet 4.6"},
	{ID: "claude-sonnet-4-6-thinking", Type: "model", DisplayName: "Claude Sonnet 4.6 (Thinking)"},
	{ID: "claude-opus-4-5-20251101", Type: "model", DisplayName: "Claude Opus 4.5"},
	{ID: "claude-opus-4-5-20251101-thinking", Type: "model", DisplayName: "Claude Opus 4.5 (Thinking)"},
	{ID: "claude-sonnet-4-5-20250929", Type: "model", DisplayName: "Claude Sonnet 4.5"},
	{ID: "claude-sonnet-4-5-20250929-thinking", Type: "model", DisplayName: "Claude Sonnet 4.5 (Thinking)"},
	{ID: "claude-haiku-4-5-20251001", Type: "model", DisplayName: "Claude Haiku 4.5"},
	{ID: "claude-haiku-4-5-20251001-thinking", Type: "model", DisplayName: "Claude Haiku 4.5 (Thinking)"},
	// 实验性 / 未经真实账号验证：Kiro 客户端近期在模型选择器里出现了 OpenAI GPT-5.6
	// 系列（截图确认，非官方文档）。请求体构造层（BuildKiroPayloadWithContext）本身
	// 模型无关，理论上可以直通这几个 modelId；但 Kiro 后端对 GPT 家族返回的响应事件流
	// 是否与 Claude 家族语义一致（tool_use/content block、stop_reason、reasoning
	// 内容位置等）完全未经验证——response 解析侧（kiroSemanticEvent 相关逻辑）目前
	// 没有做任何按模型家族的区分。上线前需要用真实 Kiro 账号实测这三个模型的非工具调用
	// 文本请求、工具调用请求、流式请求，确认解析不出现错乱后再摘掉本条注释。
	{ID: "gpt-5.6-sol", Type: "model", DisplayName: "GPT 5.6 Sol"},
	{ID: "gpt-5.6-terra", Type: "model", DisplayName: "GPT 5.6 Terra"},
	{ID: "gpt-5.6-luna", Type: "model", DisplayName: "GPT 5.6 Luna"},
}
