package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
)

// kiro_runtime_openai_bridge.go 让 Kiro 账号可以被 OpenAI 协议客户端
// （Responses API / Chat Completions API，例如 Codex CLI）访问，而不仅仅是
// Anthropic Messages 协议客户端（Claude Code）。
//
// 背景：Kiro 最初只绑定 Claude 模型，只需要 Anthropic Messages 协议入口
// （forwardKiroMessages，见 kiro_runtime.go）。Kiro 最近上线了 GPT-5.6
// 系列模型，如果只保留 Anthropic 协议入口，Codex 这类原生 OpenAI 协议客户端
// 将完全无法使用 Kiro 账号——这是一个真实的能力缺口，而非设计如此。
//
// 实现方式：不重新实现一套 Kiro 专属的 OpenAI 协议翻译器，而是复用两类已有的、
// 经过验证的基础设施：
//  1. Kiro 自身已有的 openKiroAnthropicStreamResponse：把 Kiro 原始事件流转换成
//     一个"伪造的" *http.Response，其 Body 产出标准 Anthropic SSE 字节流
//     （forwardKiroMessages 的流式分支已在使用它）。
//  2. GatewayService 已有的 Responses/ChatCompletions 桥接终端函数
//     （handleResponsesStreamingResponse 等，定义在 gateway_forward_as_responses.go /
//     gateway_forward_as_chat_completions.go）：它们只依赖“resp.Body 是 Anthropic
//     格式 SSE”这个前提，并不关心 resp 到底来自真实 Anthropic 上游还是 Kiro。
//
// 因此这里的代码只是薄薄的一层胶水：请求方向做 Responses/ChatCompletions → Anthropic
// 转换（复用 internal/pkg/apicompat），然后调用 Kiro 现成的上游函数拿到一个
// Anthropic 形状的 resp，再把这个 resp 转交给现成的 Responses/ChatCompletions
// 终端处理函数。
//
// ⚠️ 已知限制：这条路径只解决了“客户端协议格式转换”这一层——Kiro 后端本身是否
// 真的会对 GPT 模型返回与 Claude 模型语义一致的事件流（tool_use、stop_reason 等），
// 目前没有真实 Kiro GPT 账号可以验证，请参考 translator.go 顶部注释与
// /memories/session/plan.md 中的说明。

// kiroHTTPErrorAsBridgeError 是终态（非 failover）Kiro 上游错误的载体，仅用于让
// forwardKiroAsResponses/forwardKiroAsChatCompletions 在返回前区分“需要按自己协议
// 格式写错误响应”和“failover，什么都不写交给上层重试”。
type kiroHTTPErrorAsBridgeError struct {
	*kiroHTTPErrorOutcome
}

func (e *kiroHTTPErrorAsBridgeError) Error() string {
	return fmt.Sprintf("kiro upstream error: %d %s", e.StatusCode, e.Message)
}

// kiroAnthropicBridgeResult 是 bridgeKiroAsAnthropic 成功时的返回值：一个 Anthropic
// 形状 SSE 的 resp，加上请求/映射后的模型名，供调用方选择 Responses 或
// ChatCompletions 终端处理函数使用。
type kiroAnthropicBridgeResult struct {
	Resp        *http.Response
	MappedModel string
}

// bridgeKiroAsAnthropic 是 Kiro OpenAI 协议桥接的共用核心：接受一个已经转换为
// Anthropic 请求格式的 anthropicReq（分别由 Responses/ChatCompletions 转换而来），
// 完成模型映射、构造 Kiro 所需的 ParsedRequest、调用 Kiro 现有上游函数，并统一处理
// 错误分类。
//
// 返回值的错误要么是 *UpstreamFailoverError（调用方应原样返回，不写任何响应，交给
// failover 循环换账号重试），要么是 *kiroHTTPErrorAsBridgeError（调用方需要用自己
// 协议的错误格式写响应）。
func (s *GatewayService) bridgeKiroAsAnthropic(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	anthropicReq *apicompat.AnthropicRequest,
	originalModel string,
	parsed *ParsedRequest,
) (*kiroAnthropicBridgeResult, error) {
	mappedModel := originalModel
	if next := account.GetMappedModel(originalModel); next != "" {
		mappedModel = next
	}
	anthropicReq.Model = mappedModel
	// Kiro 上游统一走流式获取（与真实 Anthropic 账号在 ForwardAsResponses/
	// ForwardAsChatCompletions 中的做法一致），是否以流式返回给客户端由调用方在
	// 拿到 resp 之后自行选择终端处理函数决定。
	anthropicReq.Stream = true

	anthropicBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("marshal anthropic request for kiro: %w", err)
	}

	kiroParsed, err := ParseGatewayRequest(NewRequestBodyRef(anthropicBody), domain.PlatformAnthropic)
	if err != nil {
		return nil, fmt.Errorf("build kiro parsed request: %w", err)
	}
	if parsed != nil {
		kiroParsed.GroupID = parsed.GroupID
		kiroParsed.SessionContext = parsed.SessionContext
	}
	kiroParsed.Stream = true

	var kiroGroup *Group
	if kiroParsed.GroupID != nil && s.groupRepo != nil {
		if g, err := s.groupRepo.GetByIDLite(ctx, *kiroParsed.GroupID); err == nil {
			kiroGroup = g
		}
	}

	resp, _, err := s.openKiroAnthropicStreamResponse(ctx, account, kiroParsed, anthropicBody, mappedModel, originalModel, c.Request.Header, kiroGroup)
	if err != nil {
		var failoverErr *UpstreamFailoverError
		if errors.As(err, &failoverErr) {
			return nil, failoverErr
		}
		return nil, &kiroHTTPErrorAsBridgeError{&kiroHTTPErrorOutcome{
			StatusCode: http.StatusBadGateway,
			Message:    "Upstream request failed",
		}}
	}

	if resp.StatusCode >= 400 {
		defer func() { _ = resp.Body.Close() }()
		outcome, classifyErr := s.classifyKiroHTTPErrorAndRecordOps(ctx, resp, c, account, mappedModel, anthropicBody)
		if classifyErr != nil {
			return nil, classifyErr // *UpstreamFailoverError
		}
		return nil, &kiroHTTPErrorAsBridgeError{outcome}
	}

	return &kiroAnthropicBridgeResult{Resp: resp, MappedModel: mappedModel}, nil
}

// forwardKiroAsResponses 让 Kiro 账号可以通过 OpenAI Responses API
// （POST /v1/responses，Codex CLI 等原生 OpenAI 协议客户端使用）访问。
//
// 流程与 GatewayService.ForwardAsResponses（服务真实 Anthropic 账号）完全对称：
// 解析 Responses 请求 → 转换为 Anthropic 请求 → 转发上游 → 把响应转换回 Responses
// 格式；唯一区别是"转发上游"这一步换成了 Kiro 专用的桥接函数。
func (s *GatewayService) forwardKiroAsResponses(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	parsed *ParsedRequest,
	startTime time.Time,
) (*ForwardResult, error) {
	var responsesReq apicompat.ResponsesRequest
	if err := json.Unmarshal(body, &responsesReq); err != nil {
		return nil, fmt.Errorf("parse responses request: %w", err)
	}
	originalModel := responsesReq.Model
	clientStream := responsesReq.Stream
	reasoningEffort := ExtractResponsesReasoningEffortFromBody(body)

	anthropicReq, err := apicompat.ResponsesToAnthropicRequest(&responsesReq)
	if err != nil {
		return nil, fmt.Errorf("convert responses to anthropic: %w", err)
	}

	bridge, err := s.bridgeKiroAsAnthropic(ctx, c, account, anthropicReq, originalModel, parsed)
	if err != nil {
		var failoverErr *UpstreamFailoverError
		if errors.As(err, &failoverErr) {
			return nil, failoverErr
		}
		var bridgeErr *kiroHTTPErrorAsBridgeError
		if errors.As(err, &bridgeErr) {
			writeResponsesError(c, bridgeErr.StatusCode, "server_error", bridgeErr.Message)
			return nil, bridgeErr
		}
		writeResponsesError(c, http.StatusBadGateway, "server_error", "Upstream request failed")
		return nil, err
	}
	defer func() { _ = bridge.Resp.Body.Close() }()

	if clientStream {
		return s.handleResponsesStreamingResponse(bridge.Resp, c, originalModel, bridge.MappedModel, reasoningEffort, startTime)
	}
	return s.handleResponsesBufferedStreamingResponse(bridge.Resp, c, originalModel, bridge.MappedModel, reasoningEffort, startTime)
}

// forwardKiroAsChatCompletions 让 Kiro 账号可以通过 OpenAI Chat Completions API
// （POST /v1/chat/completions）访问。
//
// 流程与 GatewayService.ForwardAsChatCompletions（服务真实 Anthropic 账号）对称：
// ChatCompletions → Responses → Anthropic 链式转换后转发上游，响应经由现成的
// "Anthropic → Responses → ChatCompletions" 终端函数转换回 ChatCompletions 格式。
func (s *GatewayService) forwardKiroAsChatCompletions(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	parsed *ParsedRequest,
	startTime time.Time,
) (*ForwardResult, error) {
	var ccReq apicompat.ChatCompletionsRequest
	if err := json.Unmarshal(body, &ccReq); err != nil {
		return nil, fmt.Errorf("parse chat completions request: %w", err)
	}
	originalModel := ccReq.Model
	clientStream := ccReq.Stream
	includeUsage := ccReq.StreamOptions != nil && ccReq.StreamOptions.IncludeUsage
	reasoningEffort := extractCCReasoningEffortFromBody(body)

	responsesReq, err := apicompat.ChatCompletionsToResponses(&ccReq)
	if err != nil {
		return nil, fmt.Errorf("convert chat completions to responses: %w", err)
	}
	anthropicReq, err := apicompat.ResponsesToAnthropicRequest(responsesReq)
	if err != nil {
		return nil, fmt.Errorf("convert responses to anthropic: %w", err)
	}

	bridge, err := s.bridgeKiroAsAnthropic(ctx, c, account, anthropicReq, originalModel, parsed)
	if err != nil {
		var failoverErr *UpstreamFailoverError
		if errors.As(err, &failoverErr) {
			return nil, failoverErr
		}
		var bridgeErr *kiroHTTPErrorAsBridgeError
		if errors.As(err, &bridgeErr) {
			writeGatewayCCError(c, bridgeErr.StatusCode, "server_error", bridgeErr.Message)
			return nil, bridgeErr
		}
		writeGatewayCCError(c, http.StatusBadGateway, "server_error", "Upstream request failed")
		return nil, err
	}
	defer func() { _ = bridge.Resp.Body.Close() }()

	if clientStream {
		return s.handleCCStreamingFromAnthropic(bridge.Resp, c, originalModel, bridge.MappedModel, reasoningEffort, startTime, includeUsage)
	}
	return s.handleCCBufferedFromAnthropic(bridge.Resp, c, originalModel, bridge.MappedModel, reasoningEffort, startTime)
}
