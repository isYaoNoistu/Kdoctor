package logs

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"kdoctor/internal/composeutil"
	"kdoctor/internal/config"
	"kdoctor/internal/snapshot"
	dockertransport "kdoctor/internal/transport/docker"
)

type Collector struct{}

type fingerprint struct {
	ID                string
	Pattern           *regexp.Regexp
	Severity          string
	Meaning           string
	ProbableCauses    []string
	RecommendedChecks []string
}

func (Collector) Collect(ctx context.Context, env *config.Runtime, compose *snapshot.ComposeSnapshot, docker *snapshot.DockerSnapshot) *snapshot.LogSnapshot {
	if env == nil || !env.Config.Logs.Enabled {
		return nil
	}

	out := &snapshot.LogSnapshot{Collected: true}
	sourceContents := map[string]string{}

	if logDir := strings.TrimSpace(env.LogDir); logDir != "" {
		fileSources, errs := collectFileLogs(logDir, env.Config.Logs.TailLines, env.Config.Logs.LookbackMinutes)
		for source, content := range fileSources {
			sourceContents[source] = content
		}
		for _, err := range errs {
			out.Errors = append(out.Errors, err.Error())
		}
	}

	if docker != nil && docker.Available {
		names := docker.ExpectedNames
		if len(names) == 0 {
			names = composeutil.ContainerNames(compose, env.Config.Docker.ContainerNames)
		}
		since := ""
		if env.Config.Logs.LookbackMinutes > 0 {
			since = fmt.Sprintf("%dm", env.Config.Logs.LookbackMinutes)
		}
		for _, name := range names {
			content, err := dockertransport.Logs(ctx, name, env.Config.Logs.TailLines, since)
			if err != nil {
				out.Errors = append(out.Errors, err.Error())
				continue
			}
			sourceContents["docker:"+name] = content
		}
	}

	if len(sourceContents) == 0 {
		return out
	}

	out.Available = true
	for source := range sourceContents {
		out.Sources = append(out.Sources, source)
	}
	sort.Strings(out.Sources)
	out.Matches = aggregateMatches(sourceContents)
	return out
}

func collectFileLogs(logDir string, tailLines int, lookbackMinutes int) (map[string]string, []error) {
	sources := map[string]string{}
	errs := []error{}

	root := strings.TrimSpace(logDir)
	if root == "" {
		return sources, errs
	}
	info, err := os.Stat(root)
	if err != nil {
		return sources, []error{fmt.Errorf("stat log dir: %w", err)}
	}
	if !info.IsDir() {
		content, err := readTail(root, tailLines)
		if err != nil {
			return sources, []error{fmt.Errorf("read log file %s: %w", root, err)}
		}
		sources["file:"+root] = content
		return sources, errs
	}

	lookback := time.Time{}
	if lookbackMinutes > 0 {
		lookback = time.Now().Add(-time.Duration(lookbackMinutes) * time.Minute)
	}

	count := 0
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			errs = append(errs, walkErr)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !isLogLike(path) {
			return nil
		}
		if !lookback.IsZero() {
			info, err := d.Info()
			if err == nil && info.ModTime().Before(lookback) {
				return nil
			}
		}
		content, err := readTail(path, tailLines)
		if err != nil {
			errs = append(errs, fmt.Errorf("read log file %s: %w", path, err))
			return nil
		}
		sources["file:"+path] = content
		count++
		if count >= 12 {
			return io.EOF
		}
		return nil
	})
	if walkErr != nil && walkErr != io.EOF {
		errs = append(errs, walkErr)
	}
	return sources, errs
}

func readTail(path string, tailLines int) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	const maxWindow = int64(512 * 1024)
	window := info.Size()
	if window > maxWindow {
		window = maxWindow
	}

	if _, err := file.Seek(-window, io.SeekEnd); err != nil {
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return "", err
		}
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	lines = trimEmptyTail(lines)
	if tailLines <= 0 || tailLines >= len(lines) {
		return strings.Join(lines, "\n"), nil
	}
	return strings.Join(lines[len(lines)-tailLines:], "\n"), nil
}

func trimEmptyTail(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func isLogLike(path string) bool {
	lower := strings.ToLower(filepath.Base(path))
	return strings.HasSuffix(lower, ".log") || strings.HasSuffix(lower, ".out") || strings.HasSuffix(lower, ".err")
}

func aggregateMatches(sourceContents map[string]string) []snapshot.LogPatternMatch {
	type aggregate struct {
		match   snapshot.LogPatternMatch
		sources map[string]struct{}
	}

	acc := map[string]*aggregate{}
	for source, content := range sourceContents {
		lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			for _, fp := range fingerprints() {
				if !fp.Pattern.MatchString(line) {
					continue
				}
				current := acc[fp.ID]
				if current == nil {
					current = &aggregate{
						match: snapshot.LogPatternMatch{
							ID:                fp.ID,
							Pattern:           fp.Pattern.String(),
							Severity:          fp.Severity,
							Meaning:           fp.Meaning,
							Example:           line,
							ProbableCauses:    append([]string(nil), fp.ProbableCauses...),
							RecommendedChecks: append([]string(nil), fp.RecommendedChecks...),
						},
						sources: map[string]struct{}{},
					}
					acc[fp.ID] = current
				}
				current.match.Count++
				current.sources[source] = struct{}{}
			}
		}
	}

	out := make([]snapshot.LogPatternMatch, 0, len(acc))
	for _, item := range acc {
		for source := range item.sources {
			item.match.AffectedSources = append(item.match.AffectedSources, source)
		}
		sort.Strings(item.match.AffectedSources)
		out = append(out, item.match)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Severity != out[j].Severity {
			return severityRank(out[i].Severity) > severityRank(out[j].Severity)
		}
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func fingerprints() []fingerprint {
	return []fingerprint{
		newFingerprint("LOG-LEADER-NOT-AVAILABLE", `(?i)LEADER_NOT_AVAILABLE`, "fail", "partition leader is currently unavailable", []string{"controller transition is not complete", "leader broker is offline", "topic is still recovering"}, []string{"check TOP-003 leader health", "check KRF-002 active controller", "check broker registration and logs"}),
		newFingerprint("LOG-NOT-LEADER", `(?i)NOT_LEADER_OR_FOLLOWER`, "fail", "client reached a broker that is not the current leader", []string{"metadata is stale", "leader moved during failure or rebalance", "advertised.listeners returned an unreachable or wrong broker"}, []string{"check NET-003 metadata endpoint reachability", "refresh client metadata", "check topic leader distribution"}),
		newFingerprint("LOG-UNKNOWN-TOPIC", `(?i)UNKNOWN_TOPIC_OR_PARTITION`, "warn", "requested topic or partition does not exist on the cluster", []string{"topic name is wrong", "topic has not been created", "metadata is inconsistent during creation"}, []string{"check topic existence", "verify auto-create topic policy", "check controller logs"}),
		newFingerprint("LOG-OFFSET-OOR", `(?i)OffsetOutOfRange`, "warn", "consumer requested an offset outside the retained range", []string{"consumer lag exceeded retention", "retention deleted old segments", "manual offset reset is required"}, []string{"check retention policy", "check consumer lag", "reset offsets if appropriate"}),
		newFingerprint("LOG-MESSAGE-TOO-LARGE", `(?i)(MessageTooLarge|RecordTooLargeException)`, "fail", "message exceeds broker or client size limits", []string{"producer max request is larger than broker allowance", "message payload exceeds configured limits"}, []string{"check broker message.max.bytes", "check producer max.request.size", "retry with a smaller payload"}),
		newFingerprint("LOG-CONNECTION-NODE", `(?i)Connection to node .* could not be established`, "fail", "client cannot establish TCP connectivity to a broker returned by metadata", []string{"advertised.listeners is wrong", "broker port is closed", "routing or firewall blocks the returned endpoint"}, []string{"check NET-003 metadata endpoint reachability", "check CFG-006 listeners settings", "verify broker ports are exposed"}),
		newFingerprint("LOG-NODE-ASSIGNMENT", `(?i)Timed out waiting for a node assignment`, "fail", "producer could not find an assignable broker in time", []string{"metadata is stale or incomplete", "leaders are unavailable", "all returned brokers are unreachable"}, []string{"check KFK-002 broker registration", "check TOP-003 leader health", "check client metadata and network paths"}),
		newFingerprint("LOG-COORDINATOR-NOT-AVAILABLE", `(?i)Group coordinator not available`, "warn", "consumer group coordinator is not currently ready", []string{"internal topics are unhealthy", "controller is transitioning", "brokers are still recovering"}, []string{"check KFK-004 internal topics", "check controller health", "retry after cluster stabilizes"}),
		newFingerprint("LOG-COORDINATOR-LOADING", `(?i)Coordinator load in progress`, "warn", "group coordinator is still loading state", []string{"broker just started", "offset topic partition is recovering", "controller transition is in progress"}, []string{"check broker restart timeline", "check __consumer_offsets health", "retry once brokers settle"}),
		newFingerprint("LOG-NOT-CONTROLLER", `(?i)NotControllerException`, "fail", "request hit a node that is no longer the active controller", []string{"controller has changed", "controller listener is unstable", "metadata cached an old controller"}, []string{"check KRF-002 active controller", "check KRF-003 quorum majority", "check controller listener reachability"}),
		newFingerprint("LOG-NO-SPACE", `(?i)No space left on device`, "crit", "disk is full and Kafka can no longer write safely", []string{"data or metadata disk is exhausted", "retention cleanup is insufficient", "host mount is mis-sized"}, []string{"check HOST-004 disk usage", "free disk space immediately", "review retention and log segment sizing"}),
		newFingerprint("LOG-DISK-ERROR", `(?i)Disk error`, "crit", "Kafka reported a disk-level IO failure", []string{"underlying disk failure", "filesystem error", "mount path instability"}, []string{"check host and kernel logs", "verify disk health", "consider moving traffic away from the affected broker"}),
		newFingerprint("LOG-CORRUPT-RECORD", `(?i)CorruptRecordException`, "fail", "Kafka detected a corrupt record or segment", []string{"segment corruption", "unclean shutdown during IO", "storage or filesystem error"}, []string{"check broker logs around the segment", "verify disk integrity", "run partition leader and ISR checks"}),
		newFingerprint("LOG-REJECTED-EXECUTION", `(?i)RejectedExecutionException`, "warn", "Kafka thread pools are overloaded or shutting down", []string{"broker is overloaded", "broker is stopping", "system resources are insufficient"}, []string{"check host CPU and memory", "check broker restart events", "inspect request latency and backpressure"}),
		newFingerprint("LOG-OOM", `(?i)OutOfMemoryError`, "crit", "JVM ran out of memory", []string{"heap is undersized", "traffic spike exhausted memory", "memory leak or severe backlog"}, []string{"check DKR-003 OOMKilled", "review heap and container memory limits", "restart carefully after confirming memory headroom"}),
	}
}

func newFingerprint(id string, pattern string, severity string, meaning string, causes []string, checks []string) fingerprint {
	return fingerprint{
		ID:                id,
		Pattern:           regexp.MustCompile(pattern),
		Severity:          severity,
		Meaning:           meaning,
		ProbableCauses:    causes,
		RecommendedChecks: checks,
	}
}

func severityRank(severity string) int {
	switch severity {
	case "crit":
		return 4
	case "fail":
		return 3
	case "warn":
		return 2
	default:
		return 1
	}
}
