// Package composechk performs lightweight structural checks on Compose YAML files
// without calling the Docker daemon.
package composechk

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Finding is one issue or note about a Compose file.
type Finding struct {
	Level   string // "error" or "warn"
	Message string
}

// Result is the outcome of checking one file.
type Result struct {
	Path     string
	Findings []Finding
}

var (
	reService   = regexp.MustCompile(`(?m)^  ([A-Za-z0-9_-]+):\s*$`)
	reTopKey    = regexp.MustCompile(`(?m)^([A-Za-z0-9_-]+):\s*$`)
	reImage     = regexp.MustCompile(`(?m)^\s{4}image:\s*(\S+)`)
	reBuild     = regexp.MustCompile(`(?m)^\s{4}build:`)
	reLocalhost = regexp.MustCompile(`(?i)localhost|127\.0\.0\.1`)
	reLatest    = regexp.MustCompile(`(?m)^\s{4}image:\s*\S*:latest\b|^\s{4}image:\s*[A-Za-z0-9._/-]+\s*$`)
)

// CheckFile reads path and returns structural findings.
func CheckFile(path string) (Result, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Result{}, err
	}
	return Check(path, string(b)), nil
}

// Check inspects Compose YAML text.
func Check(path, content string) Result {
	r := Result{Path: path}
	if strings.TrimSpace(content) == "" {
		r.Findings = append(r.Findings, Finding{"error", "file is empty"})
		return r
	}
	top := map[string]bool{}
	for _, m := range reTopKey.FindAllStringSubmatch(content, -1) {
		top[m[1]] = true
	}
	if !top["services"] {
		r.Findings = append(r.Findings, Finding{"error", "missing top-level services key"})
		return r
	}
	services := reService.FindAllStringSubmatch(content, -1)
	if len(services) == 0 {
		r.Findings = append(r.Findings, Finding{"error", "no services found under services:"})
		return r
	}
	seen := map[string]bool{}
	for _, s := range services {
		name := s[1]
		if seen[name] {
			r.Findings = append(r.Findings, Finding{"error", "duplicate service name: " + name})
		}
		seen[name] = true
	}
	// Per-service image/build presence: naive block split
	blocks := splitServices(content)
	for name, block := range blocks {
		hasImage := reImage.MatchString(block)
		hasBuild := reBuild.MatchString(block)
		if !hasImage && !hasBuild {
			r.Findings = append(r.Findings, Finding{"error", "service " + name + " has neither image nor build"})
		}
		if reLatest.MatchString(block) && hasImage {
			r.Findings = append(r.Findings, Finding{"warn", "service " + name + " image may be unpinned (:latest or no tag)"})
		}
		if reLocalhost.MatchString(block) {
			r.Findings = append(r.Findings, Finding{"warn", "service " + name + " references localhost; inside containers that points at itself"})
		}
	}
	if !top["volumes"] {
		// only warn if any service mentions volumes: - style; skip for simplicity
	}
	if len(r.Findings) == 0 {
		r.Findings = append(r.Findings, Finding{"ok", fmt.Sprintf("%d services look structurally fine", len(services))})
	}
	return r
}

func splitServices(content string) map[string]string {
	out := map[string]string{}
	idx := strings.Index(content, "\nservices:")
	if idx < 0 {
		if strings.HasPrefix(content, "services:") {
			idx = 0
		} else {
			return out
		}
	}
	rest := content[idx:]
	// cut at next top-level key
	lines := strings.Split(rest, "\n")
	var cur string
	var buf []string
	flush := func() {
		if cur != "" {
			out[cur] = strings.Join(buf, "\n")
		}
	}
	for i, line := range lines {
		if i == 0 {
			continue
		}
		if reTopKey.MatchString(line) && !strings.HasPrefix(line, " ") {
			flush()
			cur = ""
			buf = nil
			break
		}
		if m := reService.FindStringSubmatch(line); m != nil {
			flush()
			cur = m[1]
			buf = []string{line}
			continue
		}
		if cur != "" {
			buf = append(buf, line)
		}
	}
	flush()
	return out
}
