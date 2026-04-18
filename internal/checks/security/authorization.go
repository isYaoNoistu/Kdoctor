package security

import (
	"context"
	"strings"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type AuthorizationChecker struct{}

func (AuthorizationChecker) ID() string     { return "SEC-004" }
func (AuthorizationChecker) Name() string   { return "authorization_denial" }
func (AuthorizationChecker) Module() string { return "security" }

func (AuthorizationChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	evidence := []string{}
	if bundle != nil && bundle.Logs != nil {
		for _, match := range bundle.Logs.Matches {
			if match.ID == "LOG-AUTHORIZATION" || match.ID == "LOG-AUTHENTICATION" {
				evidence = append(evidence, "日志命中="+match.ID)
			}
		}
	}
	if bundle != nil && bundle.Probe != nil && strings.Contains(strings.ToLower(bundle.Probe.Error), "authorization") {
		evidence = append(evidence, "探针错误="+bundle.Probe.Error)
	}

	if len(evidence) == 0 {
		if bundle == nil || bundle.Logs == nil || !bundle.Logs.Collected {
			return rule.NewSkip("SEC-004", "authorization_denial", "security", "当前没有可用的认证/授权证据来源")
		}
		return rule.NewPass("SEC-004", "authorization_denial", "security", "未观察到明确的认证或 ACL 拒绝证据")
	}

	result := rule.NewFail("SEC-004", "authorization_denial", "security", "已观察到认证或 ACL 拒绝证据，问题不应再简单归类为网络故障")
	result.Evidence = evidence
	result.NextActions = []string{"检查 StandardAuthorizer 与 ACL 规则", "确认客户端 principal、listener 与认证机制一致", "区分认证失败和授权拒绝，不要只排查网络"}
	return result
}
