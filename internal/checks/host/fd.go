package host

import (
	"context"
	"fmt"

	"kdoctor/internal/rule"
	"kdoctor/internal/snapshot"
	"kdoctor/pkg/model"
)

type FDChecker struct {
	WarnPct int
	CritPct int
}

func (FDChecker) ID() string     { return "HOST-008" }
func (FDChecker) Name() string   { return "fd_headroom" }
func (FDChecker) Module() string { return "host" }

func (c FDChecker) Run(_ context.Context, bundle *snapshot.Bundle) model.CheckResult {
	if bundle == nil || bundle.Host == nil || !bundle.Host.Collected || bundle.Host.FD == nil {
		if bundle == nil || bundle.Host == nil || !bundle.Host.Collected || len(bundle.Host.ContainerFD) == 0 {
			return rule.NewSkip("HOST-008", "fd_headroom", "host", "当前输入模式下没有可用的文件描述符证据")
		}
	}
	if c.WarnPct <= 0 {
		c.WarnPct = 70
	}
	if c.CritPct <= 0 {
		c.CritPct = 85
	}

	if result, ok := evaluateContainerFD(bundle.Host.ContainerFD, bundle.Host.FD); ok {
		return result
	}

	fd := bundle.Host.FD
	evidence := []string{}
	if fd.SoftLimit > 0 {
		evidence = append(evidence, fmt.Sprintf("soft_limit=%d", fd.SoftLimit))
	}
	if fd.SystemMax > 0 {
		usedPct := float64(fd.SystemUsed) * 100 / float64(fd.SystemMax)
		evidence = append(evidence, fmt.Sprintf("system_used=%d system_max=%d used_pct=%.1f", fd.SystemUsed, fd.SystemMax, usedPct))
		result := rule.NewPass("HOST-008", "fd_headroom", "host", "宿主机文件描述符余量正常")
		result.Evidence = evidence
		switch {
		case usedPct >= float64(c.CritPct) || (fd.SoftLimit > 0 && fd.SoftLimit < 32768):
			result = rule.NewFail("HOST-008", "fd_headroom", "host", "宿主机文件描述符余量已经非常紧张")
			result.Evidence = evidence
			result.NextActions = []string{"提高 Kafka 与执行环境的 ulimit -n", "检查当前文件描述符增长与 socket 抖动情况", "确认最近的连接高峰没有耗尽共享宿主机限制"}
		case usedPct >= float64(c.WarnPct) || (fd.SoftLimit > 0 && fd.SoftLimit < 65536):
			result = rule.NewWarn("HOST-008", "fd_headroom", "host", "宿主机文件描述符余量开始变紧")
			result.Evidence = evidence
			result.NextActions = []string{"在流量上升前复核 ulimit -n 和当前描述符压力", "检查连接抖动或客户端重试是否抬高了描述符占用", "为 Kafka 数据和网络负载预留更多 fd 余量"}
		}
		return result
	}

	result := rule.NewPass("HOST-008", "fd_headroom", "host", "宿主机文件描述符软限制可见，暂未表现出直接风险")
	result.Evidence = evidence
	if fd.SoftLimit > 0 && fd.SoftLimit < 65536 {
		result = rule.NewWarn("HOST-008", "fd_headroom", "host", "宿主机文件描述符软限制低于常见 Kafka 生产基线")
		result.Evidence = evidence
		result.NextActions = []string{"提高 Kafka 服务用户的 ulimit -n", "确认 broker 进程继承了预期的软硬限制", "在负载增长前复核 listener 与客户端连接扇出"}
	}
	return result
}

func evaluateContainerFD(containers []snapshot.ContainerFDStat, hostFD *snapshot.FDStats) (model.CheckResult, bool) {
	if len(containers) == 0 {
		return model.CheckResult{}, false
	}

	minSoft := uint64(0)
	usable := 0
	evidence := make([]string, 0, len(containers)+1)
	for _, item := range containers {
		if item.Error != "" {
			evidence = append(evidence, fmt.Sprintf("container=%s error=%s", item.Name, item.Error))
			continue
		}
		usable++
		evidence = append(evidence, fmt.Sprintf("container=%s soft_limit=%d hard_limit=%d", item.Name, item.SoftLimit, item.HardLimit))
		if item.SoftLimit > 0 && (minSoft == 0 || item.SoftLimit < minSoft) {
			minSoft = item.SoftLimit
		}
	}
	if hostFD != nil && hostFD.SystemMax > 0 {
		usedPct := float64(hostFD.SystemUsed) * 100 / float64(hostFD.SystemMax)
		evidence = append(evidence, fmt.Sprintf("system_used=%d system_max=%d used_pct=%.1f", hostFD.SystemUsed, hostFD.SystemMax, usedPct))
	}
	if usable == 0 || minSoft == 0 {
		result := rule.NewSkip("HOST-008", "fd_headroom", "host", "当前 Docker 场景下未能读取 Kafka 容器的文件描述符限制")
		result.Evidence = evidence
		return result, true
	}

	result := rule.NewPass("HOST-008", "fd_headroom", "host", "Kafka 容器进程文件描述符余量正常")
	result.Evidence = evidence
	switch {
	case minSoft < 32768:
		result = rule.NewFail("HOST-008", "fd_headroom", "host", "Kafka 容器进程文件描述符余量已经非常紧张")
		result.Evidence = evidence
		result.NextActions = []string{"提高 Kafka 容器或服务进程的 ulimit -n", "确认容器内 broker 进程继承了预期的 open files 限制", "在连接高峰前复核 socket 与客户端连接压力"}
	case minSoft < 65536:
		result = rule.NewWarn("HOST-008", "fd_headroom", "host", "Kafka 容器进程文件描述符余量开始变紧")
		result.Evidence = evidence
		result.NextActions = []string{"在流量上升前复核 Kafka 容器进程的 open files 限制", "确认 broker 进程与容器运行时的限制一致", "为连接增长预留更多文件描述符余量"}
	}
	return result, true
}
