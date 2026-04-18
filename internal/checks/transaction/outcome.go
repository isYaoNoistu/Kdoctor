package transaction

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type OutcomeChecker struct {
	TXProbeEnabled  bool
	TransactionalID string
}

func (OutcomeChecker) ID() string     { return "TXN-005" }
func (OutcomeChecker) Name() string   { return "transaction_outcome" }
func (OutcomeChecker) Module() string { return "transaction" }

func (c OutcomeChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	required := c.TXProbeEnabled || c.TransactionalID != ""
	if !required {
		return rule.NewSkip("TXN-005", "transaction_outcome", "transaction", "transaction outcome evidence is not required by the current profile")
	}
	if bundle == nil || bundle.Logs == nil || !bundle.Logs.Collected {
		return rule.NewSkip("TXN-005", "transaction_outcome", "transaction", "transaction outcome evidence is not available because log collection is disabled")
	}

	evidence := []string{
		fmt.Sprintf("tx_probe_enabled=%t", c.TXProbeEnabled),
		fmt.Sprintf("transactional_id=%t", c.TransactionalID != ""),
	}
	for _, match := range bundle.Logs.Matches {
		if match.ID != "LOG-TRANSACTION" {
			continue
		}
		result := rule.NewFail("TXN-005", "transaction_outcome", "transaction", "transaction logs already show commit, abort, or coordinator-side transaction errors")
		result.Evidence = append(evidence, fmt.Sprintf("log_match=%s count=%d", match.ID, match.Count))
		result.NextActions = []string{"review transaction coordinator and broker logs around the failing producer", "check transaction timeout limits and fencing-related errors", "verify __transaction_state health before retrying transactional workloads"}
		return result
	}

	result := rule.NewPass("TXN-005", "transaction_outcome", "transaction", "no explicit transaction abort or commit error signal is visible in the current log window")
	result.Evidence = evidence
	return result
}
