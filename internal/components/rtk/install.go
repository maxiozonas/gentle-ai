package rtk

import (
	"github.com/gentleman-programming/gentle-ai/internal/installcmd"
	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

// InstallCommand returns the install command sequence for the current platform.
// Mirrors engram's thin wrapper — delegates to the shared installcmd resolver
// so platform-specific logic stays centralised in internal/installcmd.
//
//   - macOS / Linux with Homebrew: brew install rtk
//   - Linux without brew: returns an error — callers must fall back to
//     DownloadLatestBinary, which fetches a prebuilt release from GitHub.
//   - Windows: returns an error — rtk hooks require a POSIX shell.
func InstallCommand(profile system.PlatformProfile) ([][]string, error) {
	return installcmd.NewResolver().ResolveComponentInstall(profile, model.ComponentRTK)
}
