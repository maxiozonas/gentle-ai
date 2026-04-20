package rtk

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gentleman-programming/gentle-ai/internal/system"
)

// Package-level vars for testability — tests replace these to avoid real subprocess calls.
var (
	cmdLookPath      = exec.LookPath
	osStat           = os.Stat
	osUserHomeDir    = os.UserHomeDir
	rtkVersionOutput = func() ([]byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return exec.CommandContext(ctx, "rtk", "--version").Output()
	}
)

// semverPattern matches version strings like "rtk 1.2.3", "v1.2.3", or "1.2.3".
var semverPattern = regexp.MustCompile(`v?(\d+\.\d+\.\d+)`)

// Available reports whether the rtk binary is reachable on the system.
// It checks PATH first, then falls back to well-known install locations:
//   - ~/.local/bin/rtk (Linux default)
//   - /opt/homebrew/bin/rtk, /usr/local/bin/rtk (macOS Homebrew)
func Available(profile system.PlatformProfile) bool {
	if _, err := cmdLookPath("rtk"); err == nil {
		return true
	}

	homeDir, err := osUserHomeDir()
	if err != nil {
		return false
	}

	// Check ~/.local/bin/rtk — the default curl install location on Linux.
	if _, err := osStat(filepath.Join(homeDir, ".local", "bin", "rtk")); err == nil {
		return true
	}

	// Check well-known Homebrew prefixes for macOS (arm64 and x86).
	// RTK may be installed via brew but not yet in the shell PATH.
	if profile.OS == "darwin" || profile.PackageManager == "brew" {
		for _, path := range []string{
			"/opt/homebrew/bin/rtk",
			"/usr/local/bin/rtk",
		} {
			if _, err := osStat(path); err == nil {
				return true
			}
		}
	}

	return false
}

// VerifyInstallation runs `rtk --version` to confirm the binary is functional.
// Returns an error if the binary cannot be executed or produces unexpected output.
func VerifyInstallation() error {
	out, err := rtkVersionOutput()
	if err != nil {
		return fmt.Errorf("rtk --version failed: %w", err)
	}

	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return fmt.Errorf("rtk --version returned empty output")
	}

	if !semverPattern.MatchString(trimmed) {
		return fmt.Errorf("rtk --version returned unexpected output: %s", trimmed)
	}

	return nil
}

// VerifyVersion runs `rtk --version` and returns the parsed semver string.
// The version string is the numeric portion only (e.g., "1.2.3").
func VerifyVersion() (string, error) {
	out, err := rtkVersionOutput()
	if err != nil {
		return "", fmt.Errorf("rtk --version failed: %w", err)
	}

	trimmed := strings.TrimSpace(string(out))
	matches := semverPattern.FindStringSubmatch(trimmed)
	if len(matches) < 2 {
		return "", fmt.Errorf("rtk --version returned unexpected output: %s", trimmed)
	}

	return matches[1], nil
}
