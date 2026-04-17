package lint

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type ProcessRolesChecker struct {
	RequireController bool
	RequireBroker     bool
}

func (ProcessRolesChecker) ID() string     { return "CFG-004" }
func (ProcessRolesChecker) Name() string   { return "process_roles_legality" }
func (ProcessRolesChecker) Module() string { return "lint" }

func (c ProcessRolesChecker) Run(_ context.Context, snap *snapshot.Bundle) model.CheckResult {
	services := kafkaServices(getCompose(snap))
	if len(services) == 0 {
		return rule.NewSkip("CFG-004", "process_roles_legality", "lint", "compose Kafka services not available")
	}

	allowed := map[string]struct{}{
		"broker":     {},
		"controller": {},
	}
	evidence := []string{}
	for _, service := range services {
		if len(service.ProcessRoles) == 0 {
			result := rule.NewFail("CFG-004", "process_roles_legality", "lint", "process.roles missing in Kafka service")
			result.Evidence = []string{fmt.Sprintf("service=%s", service.ServiceName)}
			return result
		}
		for _, role := range service.ProcessRoles {
			if _, ok := allowed[role]; !ok {
				result := rule.NewFail("CFG-004", "process_roles_legality", "lint", "process.roles contains unsupported role")
				result.Evidence = []string{fmt.Sprintf("service=%s role=%s", service.ServiceName, role)}
				return result
			}
		}
		if c.RequireBroker && !contains(service.ProcessRoles, "broker") {
			result := rule.NewFail("CFG-004", "process_roles_legality", "lint", "process.roles is missing broker role")
			result.Evidence = []string{fmt.Sprintf("service=%s roles=%v", service.ServiceName, service.ProcessRoles)}
			return result
		}
		if c.RequireController && !contains(service.ProcessRoles, "controller") {
			result := rule.NewFail("CFG-004", "process_roles_legality", "lint", "process.roles is missing controller role")
			result.Evidence = []string{fmt.Sprintf("service=%s roles=%v", service.ServiceName, service.ProcessRoles)}
			return result
		}
		evidence = append(evidence, fmt.Sprintf("%s roles=%v", service.ServiceName, service.ProcessRoles))
	}

	result := rule.NewPass("CFG-004", "process_roles_legality", "lint", "process.roles are legal for detected Kafka services")
	result.Evidence = evidence
	return result
}
