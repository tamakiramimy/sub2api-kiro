package service

import (
	"context"
	"time"

	kirocooldown "github.com/Wei-Shaw/sub2api/internal/kiro/cooldown"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// defaultKiroStreamKeepalive 是 Kiro 流式响应默认的 keepalive 间隔（未显式配置时使用）。
const defaultKiroStreamKeepalive = 25 * time.Second

// kiroConservativeFallbackBillingModel 是 Kiro 账号在上游返回 auto 等无法定价模型时
// 使用的保守计费兜底模型（费用取该模型的价格，避免漏计费）。
const kiroConservativeFallbackBillingModel = "claude-opus-4-6"

type kiroCooldownRecoveryAttemptedKeyType struct{}

// kiroCooldownRecoveryAttemptedKey 用于在单次请求的重试链路中标记"已经尝试过冷却池恢复"，
// 避免同一请求在多次账号切换重试时重复触发冷却池恢复逻辑。
var kiroCooldownRecoveryAttemptedKey = kiroCooldownRecoveryAttemptedKeyType{}

// streamKeepaliveIntervalForAccount 返回该账号流式响应应使用的 keepalive 间隔；
// Kiro 平台优先使用 gateway.kiro_stream_keepalive_interval，未配置时回退到默认 25s，
// 其他平台沿用通用的 gateway.stream_keepalive_interval。
func (s *GatewayService) streamKeepaliveIntervalForAccount(account *Account) time.Duration {
	if account != nil && account.Platform == PlatformKiro {
		if s != nil && s.cfg != nil && s.cfg.Gateway.KiroStreamKeepaliveInterval > 0 {
			return time.Duration(s.cfg.Gateway.KiroStreamKeepaliveInterval) * time.Second
		}
		return defaultKiroStreamKeepalive
	}
	if s != nil && s.cfg != nil && s.cfg.Gateway.StreamKeepaliveInterval > 0 {
		return time.Duration(s.cfg.Gateway.StreamKeepaliveInterval) * time.Second
	}
	return 0
}

// tryRecoverKiroCooldownPool 在候选 Kiro 账号池全部处于瞬时(429)冷却时，尝试主动清除
// 最早到期的一个冷却状态以恢复调度，避免因短暂限流导致整个分组不可用。
//
// 已接入 gateway_scheduling.go 的 SelectAccountWithLoadAwareness 候选筛选兜底路径
// （候选为空时触发一次恢复重试），并配合 isKiroRuntimeSchedulable 在候选筛选阶段跳过
// 处于活跃冷却状态的 Kiro 账号。
func (s *GatewayService) tryRecoverKiroCooldownPool(ctx context.Context, accounts []Account, requestedModel string, excludedIDs map[int64]struct{}, allowMixedScheduling bool) bool {
	if s == nil || s.kiroCooldownStore == nil || ctx.Value(kiroCooldownRecoveryAttemptedKey) == true {
		return false
	}
	tokenKeys := s.kiroTransientCooldownRecoveryKeys(ctx, accounts, requestedModel, excludedIDs, allowMixedScheduling)
	if len(tokenKeys) == 0 {
		return false
	}
	cleared, err := s.kiroCooldownStore.ClearEarliestTransientCooldown(ctx, tokenKeys)
	if err != nil {
		logger.LegacyPrintf("service.gateway", "Kiro cooldown pool recovery failed: %v", err)
		return false
	}
	if cleared {
		logger.LegacyPrintf("service.gateway", "Kiro cooldown pool recovery cleared one transient cooldown")
	}
	return cleared
}

// kiroTransientCooldownRecoveryKeys 收集候选账号中"全部处于 429 瞬时冷却"场景下可供
// ClearEarliestTransientCooldown 尝试恢复的 token key 列表；只要出现任一账号不满足
// 全员瞬时冷却的条件，就返回空列表（放弃本轮恢复，交由正常调度/冷却过期处理）。
func (s *GatewayService) kiroTransientCooldownRecoveryKeys(ctx context.Context, accounts []Account, requestedModel string, excludedIDs map[int64]struct{}, allowMixedScheduling bool) []string {
	tokenKeys := make([]string, 0, len(accounts))
	eligible := 0
	for i := range accounts {
		acc := &accounts[i]
		if acc == nil || acc.Platform != PlatformKiro || acc.Type != AccountTypeOAuth {
			if allowMixedScheduling {
				continue
			}
			return nil
		}
		if _, excluded := excludedIDs[acc.ID]; excluded {
			continue
		}
		if !acc.IsSchedulable() {
			continue
		}
		if requestedModel != "" && !s.isModelSupportedByAccountWithContext(ctx, acc, requestedModel) {
			continue
		}
		if !s.isAccountSchedulableForQuota(acc) ||
			!s.isAccountSchedulableForWindowCost(ctx, acc, false) ||
			!s.isAccountSchedulableForRPM(ctx, acc, false) {
			continue
		}
		eligible++
		state, err := s.getKiroCooldownState(ctx, buildKiroAccountKey(acc))
		if err != nil || state == nil || !state.Active {
			return nil
		}
		if state.Reason != kirocooldown.CooldownReason429 {
			return nil
		}
		tokenKeys = append(tokenKeys, buildKiroAccountKey(acc))
	}
	if eligible == 0 {
		return nil
	}
	return tokenKeys
}

// shouldUseKiroConservativeBillingFallback 判断是否应对本次结果使用保守计费兜底
// （仅 Kiro 账号，用于上游返回 auto 等无法直接定价的模型名时）。
func shouldUseKiroConservativeBillingFallback(result *ForwardResult, billingModel string, opts *recordUsageOpts) bool {
	if result == nil {
		return false
	}
	return opts != nil && opts.IsKiroAccount
}

// calculateKiroConservativeTokenCost 使用保守兜底模型（kiroConservativeFallbackBillingModel）
// 计算本次用量费用，避免因模型名无法定价而漏计费。
func (s *GatewayService) calculateKiroConservativeTokenCost(tokens UsageTokens, multiplier float64) *CostBreakdown {
	if s == nil || s.billingService == nil {
		return nil
	}
	cost, err := s.billingService.CalculateCost(kiroConservativeFallbackBillingModel, tokens, multiplier)
	if err != nil {
		return nil
	}
	return cost
}
