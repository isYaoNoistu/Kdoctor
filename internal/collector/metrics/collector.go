package metrics

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/config"
	"kdoctor/internal/snapshot"
)

type Collector struct{}

func (Collector) Collect(ctx context.Context, env *config.Runtime, compose *snapshot.ComposeSnapshot) *snapshot.MetricsSnapshot {
	if env == nil || !env.EnableJMX {
		return nil
	}

	out := &snapshot.MetricsSnapshot{
		Collected: true,
		Path:      env.JMXPath,
	}
	endpoints := discoverEndpoints(env, compose)
	if len(endpoints) == 0 {
		out.Errors = append(out.Errors, "未配置或未推断出可用的 JMX 指标端点")
		return out
	}

	timeout := env.JMXScrapeTimeout
	if timeout <= 0 || (env.JMXTimeout > 0 && env.JMXTimeout < timeout) {
		timeout = env.JMXTimeout
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	client := &http.Client{Timeout: timeout}
	for _, endpoint := range endpoints {
		status := fetchEndpoint(ctx, client, env.JMXPath, endpoint)
		out.Endpoints = append(out.Endpoints, status)
		if status.Reachable {
			out.Available = true
		} else if status.Error != "" {
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %s", status.Address, status.Error))
		}
	}

	return out
}

func fetchEndpoint(ctx context.Context, client *http.Client, path string, endpoint endpointRef) snapshot.MetricsEndpointStatus {
	status := snapshot.MetricsEndpointStatus{
		Name:    endpoint.Name,
		Address: endpoint.Address,
	}

	targetURL := buildURL(endpoint.Address, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		status.Error = fmt.Sprintf("构造 JMX 请求失败: %v", err)
		return status
	}

	startedAt := time.Now()
	resp, err := client.Do(req)
	status.DurationMs = time.Since(startedAt).Milliseconds()
	if err != nil {
		status.Error = err.Error()
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		status.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return status
	}
	if dateHeader := strings.TrimSpace(resp.Header.Get("Date")); dateHeader != "" {
		if parsed, err := http.ParseTime(dateHeader); err == nil {
			status.ServerTimeUnix = parsed.Unix()
		}
	}

	metrics, err := parsePrometheus(resp.Body)
	if err != nil {
		status.Error = fmt.Sprintf("解析指标失败: %v", err)
		return status
	}

	status.Reachable = true
	status.Metrics = metrics
	return status
}

type endpointRef struct {
	Name    string
	Address string
}

func discoverEndpoints(env *config.Runtime, compose *snapshot.ComposeSnapshot) []endpointRef {
	if env == nil {
		return nil
	}

	out := make([]endpointRef, 0)
	seen := map[string]struct{}{}
	for _, raw := range env.JMXEndpoints {
		address := strings.TrimSpace(raw)
		if address == "" {
			continue
		}
		if _, ok := seen[address]; ok {
			continue
		}
		seen[address] = struct{}{}
		out = append(out, endpointRef{Name: address, Address: address})
	}
	if len(out) > 0 {
		return out
	}

	for _, service := range composeutil.KafkaServices(compose) {
		port := firstNonEmpty(
			service.Environment["KAFKA_JMX_PORT"],
			service.Environment["JMX_PORT"],
			service.Environment["KAFKA_CFG_JMX_PORT"],
		)
		if strings.TrimSpace(port) == "" {
			continue
		}

		host := firstNonEmpty(
			service.Environment["KAFKA_JMX_HOSTNAME"],
			service.Environment["JMX_HOSTNAME"],
		)
		if host == "" {
			host = inferHost(service, env.SelectedProfile.ExecutionView)
		}
		if host == "" {
			continue
		}

		address := fmt.Sprintf("%s:%s", host, strings.TrimSpace(port))
		if _, ok := seen[address]; ok {
			continue
		}
		seen[address] = struct{}{}
		name := service.ServiceName
		if strings.TrimSpace(service.ContainerName) != "" {
			name = service.ContainerName
		}
		out = append(out, endpointRef{Name: name, Address: address})
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Address < out[j].Address
	})
	return out
}

func inferHost(service composeutil.KafkaService, executionView string) string {
	if strings.EqualFold(strings.TrimSpace(service.NetworkMode), "host") {
		return "127.0.0.1"
	}

	switch strings.ToLower(strings.TrimSpace(executionView)) {
	case "docker-container":
		if strings.TrimSpace(service.ContainerName) != "" {
			return strings.TrimSpace(service.ContainerName)
		}
		return strings.TrimSpace(service.ServiceName)
	case "internal", "host-network", "bastion":
		if host := hostFromListeners(service.Environment["KAFKA_CFG_LISTENERS"], "INTERNAL"); host != "" {
			return host
		}
	case "external":
		if host := hostFromListeners(service.Environment["KAFKA_CFG_ADVERTISED_LISTENERS"], "EXTERNAL"); host != "" {
			return host
		}
	}
	if strings.TrimSpace(service.ContainerName) != "" {
		return strings.TrimSpace(service.ContainerName)
	}
	return strings.TrimSpace(service.ServiceName)
}

func hostFromListeners(raw string, preferred string) string {
	listeners, err := composeutil.ParseListeners(raw)
	if err != nil || len(listeners) == 0 {
		return ""
	}
	for name, listener := range listeners {
		if strings.EqualFold(name, preferred) {
			host := strings.TrimSpace(listener.Host)
			switch host {
			case "", "0.0.0.0", "::":
				return "127.0.0.1"
			default:
				return host
			}
		}
	}
	return ""
}

func parsePrometheus(reader io.Reader) (map[string]float64, error) {
	out := map[string]float64{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		name := normalizeMetricName(fields[0])
		value, err := strconv.ParseFloat(fields[len(fields)-1], 64)
		if err != nil {
			continue
		}
		out[name] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func normalizeMetricName(raw string) string {
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, "{"); idx >= 0 {
		raw = raw[:idx]
	}
	return strings.ToLower(raw)
}

func buildURL(address string, path string) string {
	address = strings.TrimSpace(address)
	if strings.HasPrefix(address, "http://") || strings.HasPrefix(address, "https://") {
		if path == "" {
			return address
		}
		u, err := url.Parse(address)
		if err != nil {
			return address
		}
		if strings.TrimSpace(u.Path) == "" || u.Path == "/" {
			u.Path = path
		}
		return u.String()
	}
	if strings.TrimSpace(path) == "" {
		path = "/metrics"
	}
	return "http://" + address + path
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
