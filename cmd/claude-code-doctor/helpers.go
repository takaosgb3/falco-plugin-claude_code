package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// findFalcoBinary mirrors the resolution order used by the Level 3 integration
// tests: ~/bin/falco first (local install), then PATH.
func findFalcoBinary() string {
	if home, err := os.UserHomeDir(); err == nil {
		candidate := filepath.Join(home, "bin", "falco")
		if st, err := os.Stat(candidate); err == nil && st.Mode()&0o111 != 0 {
			return candidate
		}
	}
	if path, err := exec.LookPath("falco"); err == nil {
		return path
	}
	return ""
}

// falcoVersion runs `falco --version` and returns the first line.
func falcoVersion(falcoBin string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, falcoBin, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	return line, nil
}

// runFalcoList runs `falco -L -c <config> --disable-source syscall -U` and
// returns the combined stdout/stderr. The -L mode lists loaded plugins / rules
// without entering the event loop, so it terminates quickly.
func runFalcoList(falcoBin, configPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, falcoBin,
		"-L",
		"-c", configPath,
		"--disable-source", "syscall",
		"-U",
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// defaultPluginPath returns the OS/arch-suffixed plugin filename, located in
// the current working directory. Mirrors the Makefile's BINARY logic.
func defaultPluginPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return filepath.Join(cwd, "libclaude-code-plugin-darwin-arm64.dylib")
		}
		return filepath.Join(cwd, "libclaude-code-plugin-darwin-amd64.dylib")
	case "linux":
		if runtime.GOARCH == "arm64" {
			return filepath.Join(cwd, "libclaude-code-plugin-linux-arm64.so")
		}
		return filepath.Join(cwd, "libclaude-code-plugin-linux-amd64.so")
	}
	return ""
}

// defaultEventsPath returns ~/.claude/security/events.jsonl. We deliberately
// don't expand here so callers can show the unresolved tilde to operators.
func defaultEventsPath() string {
	return "~/.claude/security/events.jsonl"
}

// expandPath handles ~/ expansion and rejects ".." traversal (matches HL-011
// in the hook logger).
func expandPath(p string) (string, error) {
	if strings.Contains(p, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", p)
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}

// repoRoot tries to locate the repository root by walking upward from the
// current working directory until it finds a go.mod. Returns ("", false) if
// not found within 6 levels.
func repoRoot() (string, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false
	}
	dir := cwd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

// readLastJSONLine reads the file from the end and returns the last
// non-empty newline-delimited line. Streams from the tail to avoid loading
// large files entirely.
func readLastJSONLine(path string) (string, error) {
	f, err := os.Open(path) // #nosec G304 — operator-supplied diagnostic path
	if err != nil {
		return "", err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return "", err
	}
	const chunk = 16 * 1024
	size := st.Size()
	if size == 0 {
		return "", errors.New("file is empty")
	}
	var (
		buf       []byte
		readSoFar int64
	)
	for readSoFar < size {
		toRead := int64(chunk)
		if toRead > size-readSoFar {
			toRead = size - readSoFar
		}
		seekFrom := size - readSoFar - toRead
		if _, err := f.Seek(seekFrom, io.SeekStart); err != nil {
			return "", err
		}
		tmp := make([]byte, toRead)
		if _, err := io.ReadFull(f, tmp); err != nil {
			return "", err
		}
		buf = append(tmp, buf...)
		readSoFar += toRead
		// We need at least 2 newlines (one separating the last line from EOF
		// and one starting it) — allow trailing newline.
		// Strip trailing empty lines:
		trimmed := strings.TrimRight(string(buf), "\r\n")
		idx := strings.LastIndex(trimmed, "\n")
		if idx >= 0 {
			return trimmed[idx+1:], nil
		}
		if readSoFar >= size {
			// only one line in the file
			return trimmed, nil
		}
	}
	return strings.TrimRight(string(buf), "\r\n"), nil
}

// parseDuration is like time.ParseDuration but rejects unitless integers per
// OPS-005 ("integer alone is invalid; require s/m/h units").
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty duration")
	}
	// reject pure integer (e.g. "900") — OPS-005 requires a unit
	allDigits := true
	for _, r := range s {
		if !unicode.IsDigit(r) {
			allDigits = false
			break
		}
	}
	if allDigits {
		// also accept "0" as zero
		if v, err := strconv.Atoi(s); err == nil && v == 0 {
			return 0, nil
		}
		return 0, fmt.Errorf("duration %q has no unit; use s/m/h (e.g. 15m, 1h30m, 30s)", s)
	}
	return time.ParseDuration(s)
}

// excerpt clips long strings for diagnostic output. Adds an ellipsis when
// truncated.
func excerpt(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + " …(truncated)"
}
