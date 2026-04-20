package rtk

import (
	"fmt"

	"github.com/gentleman-programming/gentle-ai/internal/system"
)

// CommandSequence represents an ordered list of commands to run in sequence.
// Each inner slice is a single command with its arguments.
// Alias matches the installcmd.CommandSequence type used throughout gentle-ai.
type CommandSequence = [][]string

// InstallCommand returns the platform-appropriate command sequence for installing RTK.
//
//   - macOS with Homebrew: brew install rtk
//   - Linux (apt/pacman/dnf): curl download of RTK's official install script + execute
//   - Windows: returns error (RTK hooks require Unix)
func InstallCommand(profile system.PlatformProfile) (CommandSequence, error) {
	switch profile.PackageManager {
	case "brew":
		return CommandSequence{
			{"brew", "install", "rtk"},
		}, nil
	case "apt", "pacman", "dnf":
		// Use RTK's official install script which handles download, extraction,
		// and PATH setup for all Linux architectures. The script is fetched
		// via curl and piped to sh — this matches the RTK recommended install
		// method documented at https://github.com/rtk-ai/rtk.
		return CommandSequence{
			{"sh", "-c", "curl -fsSL https://raw.githubusercontent.com/rtk-ai/rtk/refs/heads/master/install.sh | sh"},
		}, nil
	default:
		return nil, fmt.Errorf(
			"unsupported platform for rtk: os=%q pm=%q (RTK hooks require Unix)",
			profile.OS, profile.PackageManager,
		)
	}
}

// ShouldInstall reports whether the RTK component should be installed.
// RTK is always optional — it installs only when explicitly enabled.
func ShouldInstall(enabled bool) bool {
	return enabled
}
