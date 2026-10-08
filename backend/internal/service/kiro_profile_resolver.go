package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	kiropkg "github.com/Wei-Shaw/sub2api/internal/kiro"
	"github.com/Wei-Shaw/sub2api/internal/pkg/httpclient"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// kiroListProfilesPageSize 上游对 maxResults 的硬上限是 10，传更大会 400。
	kiroListProfilesPageSize = 10
	kiroMaxProfilePages      = 5
	// kiroProfileResolveNegCacheTTL Builder ID 等不支持 ListAvailableProfiles 的账号
	// 会持续解析失败，负缓存避免每次用量查询/刷新都重复打上游。
	kiroProfileResolveNegCacheTTL = 10 * time.Minute
)

type kiroAvailableProfile struct {
	ARN  string `json:"arn"`
	Name string `json:"profileName"`
}

type kiroListProfilesResponse struct {
	Profiles  []kiroAvailableProfile `json:"profiles"`
	NextToken string                 `json:"nextToken"`
}

var kiroProfileResolveNegCache sync.Map // accountID(int64) -> time.Time

// resolveKiroProfileArn 返回账号可用的 Kiro profileArn。
// AWS 自 2026-08 起要求 getUsageLimits 必须携带有效 profileArn（缺失或与账号
// 不匹配均返回 403 "User is not authorized to make this call"），因此凭证中
// 没有 profile_arn（或调用方判定其已失效并指定 forceRefresh）时，通过
// ListAvailableProfiles 用账号自身 token 动态解析并回写凭证。
func resolveKiroProfileArn(ctx context.Context, repo AccountRepository, account *Account, region, token string, forceRefresh bool) (string, error) {
	if account == nil {
		return "", fmt.Errorf("account is nil")
	}
	if strings.TrimSpace(token) == "" {
		return "", fmt.Errorf("no access token available")
	}
	if !forceRefresh {
		if stored := strings.TrimSpace(account.GetCredential("profile_arn")); stored != "" {
			return stored, nil
		}
		if ts, ok := kiroProfileResolveNegCache.Load(account.ID); ok {
			if t, ok := ts.(time.Time); ok && time.Since(t) < kiroProfileResolveNegCacheTTL {
				return "", fmt.Errorf("kiro profile resolution recently failed, skipping retry")
			}
		}
	}

	profiles, err := listKiroAvailableProfiles(ctx, account, region, token)
	if err != nil {
		kiroProfileResolveNegCache.Store(account.ID, time.Now())
		return "", err
	}
	arn := selectKiroProfileARN(profiles, region)
	if arn == "" {
		kiroProfileResolveNegCache.Store(account.ID, time.Now())
		return "", fmt.Errorf("no kiro profile available in region %s", region)
	}
	kiroProfileResolveNegCache.Delete(account.ID)

	if repo != nil && strings.TrimSpace(account.GetCredential("profile_arn")) != arn {
		credentials := MergeCredentials(account.Credentials, map[string]any{"profile_arn": arn})
		if err := persistAccountCredentials(ctx, repo, account, credentials); err != nil {
			// 解析结果本次仍然可用，仅记录回写失败
			logger.L().Warn("kiro.resolve_profile_arn persist failed",
				zap.Int64("account_id", account.ID),
				zap.Error(err),
			)
		}
	}
	return arn, nil
}

// selectKiroProfileARN 优先选择与数据面 region 匹配的 profile，否则取第一个。
func selectKiroProfileARN(profiles []kiroAvailableProfile, region string) string {
	region = strings.TrimSpace(region)
	for _, p := range profiles {
		arn := strings.TrimSpace(p.ARN)
		if arn == "" {
			continue
		}
		if region == "" || strings.Contains(arn, ":codewhisperer:"+region+":") {
			return arn
		}
	}
	for _, p := range profiles {
		if arn := strings.TrimSpace(p.ARN); arn != "" {
			return arn
		}
	}
	return ""
}

func listKiroAvailableProfiles(ctx context.Context, account *Account, region, token string) ([]kiroAvailableProfile, error) {
	endpoint := resolveKiroRuntimeEndpoint(region) + "/ListAvailableProfiles"
	client, err := httpclient.GetClient(httpclient.Options{
		ProxyURL:           accountProxyURL(account),
		Timeout:            30 * time.Second,
		ValidateResolvedIP: true,
		AllowPrivateHosts:  isLoopbackEndpoint(endpoint),
	})
	if err != nil {
		return nil, fmt.Errorf("create kiro profile client failed: %w", err)
	}

	var profiles []kiroAvailableProfile
	nextToken := ""
	for page := 0; page < kiroMaxProfilePages; page++ {
		payload := map[string]any{"maxResults": kiroListProfilesPageSize}
		if nextToken != "" {
			payload["nextToken"] = nextToken
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("build kiro profile request failed: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create kiro profile request failed: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		setKiroRuntimeHeaders(req, account, token)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("kiro profile request failed: %w", err)
		}
		respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read kiro profile response failed: %w", readErr)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, &kiroUsageHTTPError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(respBody))}
		}

		var parsed kiroListProfilesResponse
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			return nil, fmt.Errorf("decode kiro profile response failed: %w", err)
		}
		profiles = append(profiles, parsed.Profiles...)
		nextToken = strings.TrimSpace(parsed.NextToken)
		if nextToken == "" {
			break
		}
	}
	return profiles, nil
}

// setKiroRuntimeHeaders 设置 Kiro 数据面请求的统一运行时头（用量查询、
// profile 解析、聊天共用的指纹形态）。
func setKiroRuntimeHeaders(req *http.Request, account *Account, token string) {
	if req == nil {
		return
	}
	accountKey := buildKiroAccountKey(account)
	machineID := buildKiroMachineID(account)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	req.Header.Set("User-Agent", kiropkg.BuildRuntimeUserAgent(accountKey, machineID))
	req.Header.Set("X-Amz-User-Agent", kiropkg.BuildRuntimeAmzUserAgent(accountKey, machineID))
	req.Header.Set("x-amzn-kiro-agent-mode", "vibe")
	req.Header.Set("x-amzn-codewhisperer-optout", "true")
	req.Header.Set("Amz-Sdk-Request", "attempt=1; max=3")
	req.Header.Set("Amz-Sdk-Invocation-Id", uuid.NewString())

	if account == nil {
		return
	}
	applyKiroConditionalHeaders(req, account)
}
