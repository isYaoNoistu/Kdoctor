package docker

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type MemoryPlanningChecker struct{}

func (MemoryPlanningChecker) ID() string     { return "DKR-007" }
func (MemoryPlanningChecker) Name() string   { return "container_memory_headroom" }
func (MemoryPlanningChecker) Module() string { return "docker" }

func (MemoryPlanningChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Compose == nil {
		return rule.NewSkip("DKR-007", "container_memory_headroom", "docker", "当前没有 compose 快照，无法评估容器内存与 JVM 堆规划")
	}

	services := composeutil.KafkaServices(bundle.Compose)
	if len(services) == 0 {
		return rule.NewSkip("DKR-007", "container_memory_headroom", "docker", "compose 中没有识别到 Kafka 服务")
	}

	warnings := 0
	failures := 0
	evidence := []string{}
	for _, service := range services {
		memLimitBytes := parseHumanBytes(service.NetworkMode, service.Environment["MEM_LIMIT"])
		if memLimitBytes == 0 {
			memLimitBytes = parseHumanBytes("", service.Environment["mem_limit"])
		}
		if memLimitBytes == 0 {
			memLimitBytes = parseHumanBytes("", service.MemLimit)
		}
		xmxBytes := parseHeapXmx(service.Environment["KAFKA_HEAP_OPTS"])
		evidence = append(evidence, fmt.Sprintf("service=%s mem_limit=%s heap_opts=%s", service.ServiceName, strings.TrimSpace(service.MemLimit), strings.TrimSpace(service.Environment["KAFKA_HEAP_OPTS"])))
		if memLimitBytes == 0 || xmxBytes == 0 {
			continue
		}
		ratio := float64(xmxBytes) / float64(memLimitBytes)
		evidence = append(evidence, fmt.Sprintf("service=%s heap_to_limit_ratio=%.2f", service.ServiceName, ratio))
		switch {
		case ratio >= 0.9:
			failures++
		case ratio >= 0.8:
			warnings++
		}
	}

	result := rule.NewPass("DKR-007", "container_memory_headroom", "docker", "容器内存限制与 JVM 堆规划保留了基本余量")
	result.Evidence = evidence
	if failures > 0 {
		result = rule.NewFail("DKR-007", "container_memory_headroom", "docker", "部分 Kafka 容器的 JVM 堆已逼近容器内存上限，存在高压或 OOM 风险")
		result.Evidence = evidence
		result.NextActions = []string{"降低 Xmx 或提升容器内存限制", "为堆外内存、页缓存和本地缓冲区保留余量", "结合 DKR-003 和 JVM/GC 指标一起确认内存压力"}
		return result
	}
	if warnings > 0 {
		result = rule.NewWarn("DKR-007", "container_memory_headroom", "docker", "部分 Kafka 容器的 JVM 堆与容器内存限制过于接近，建议提前扩余量")
		result.Evidence = evidence
		result.NextActions = []string{"为 JVM 堆和容器限制保留更充足的余量", "避免把 mem_limit 大部分直接给 Xmx", "在流量高峰前复核内存与 GC 压力"}
	}
	return result
}

var heapPattern = regexp.MustCompile(`-Xmx([0-9]+)([kKmMgG])`)

func parseHeapXmx(input string) int64 {
	match := heapPattern.FindStringSubmatch(strings.TrimSpace(input))
	if len(match) != 3 {
		return 0
	}
	value, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return 0
	}
	switch strings.ToLower(match[2]) {
	case "k":
		return value * 1024
	case "m":
		return value * 1024 * 1024
	case "g":
		return value * 1024 * 1024 * 1024
	default:
		return 0
	}
}

func parseHumanBytes(_ string, input string) int64 {
	value := strings.TrimSpace(strings.ToLower(filepath.Clean(strings.TrimSpace(input))))
	value = strings.Trim(value, ".")
	if value == "" {
		return 0
	}
	multiplier := int64(1)
	switch {
	case strings.HasSuffix(value, "g"), strings.HasSuffix(value, "gb"):
		multiplier = 1024 * 1024 * 1024
		value = strings.TrimSuffix(strings.TrimSuffix(value, "gb"), "g")
	case strings.HasSuffix(value, "m"), strings.HasSuffix(value, "mb"):
		multiplier = 1024 * 1024
		value = strings.TrimSuffix(strings.TrimSuffix(value, "mb"), "m")
	case strings.HasSuffix(value, "k"), strings.HasSuffix(value, "kb"):
		multiplier = 1024
		value = strings.TrimSuffix(strings.TrimSuffix(value, "kb"), "k")
	}
	number, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return int64(number * float64(multiplier))
}
