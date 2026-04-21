package rtk

import (
	"reflect"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/system"
)

func TestInstallCommand_MacOSBrew(t *testing.T) {
	profile := system.PlatformProfile{
		OS:             "darwin",
		PackageManager: "brew",
		Supported:      true,
	}

	cmds, err := InstallCommand(profile)
	if err != nil {
		t.Fatalf("InstallCommand(darwin/brew) error = %v", err)
	}

	want := [][]string{{"brew", "install", "rtk"}}
	if !reflect.DeepEqual(cmds, want) {
		t.Errorf("InstallCommand(darwin/brew) = %v, want %v", cmds, want)
	}
}

func TestInstallCommand_LinuxUsesDirectDownload(t *testing.T) {
	// On Linux without Homebrew, InstallCommand must NOT return a
	// `curl | sh` command sequence. Callers are expected to fall back to
	// rtk.DownloadLatestBinary instead. This matches the project's stance
	// against pipe-to-shell installs (see resolveKimiInstall for the
	// original declaration of that policy).
	for _, pm := range []string{"apt", "pacman", "dnf"} {
		t.Run(pm, func(t *testing.T) {
			profile := system.PlatformProfile{
				OS:             "linux",
				PackageManager: pm,
				Supported:      true,
			}
			cmds, err := InstallCommand(profile)
			if err == nil {
				t.Fatalf("InstallCommand(linux/%s) should error to force the download path, got cmds=%v", pm, cmds)
			}
			if !strings.Contains(err.Error(), "DownloadLatestBinary") {
				t.Errorf("error should mention DownloadLatestBinary fallback, got %q", err.Error())
			}
		})
	}
}

func TestInstallCommand_UnsupportedPlatform(t *testing.T) {
	profile := system.PlatformProfile{
		OS:             "windows",
		PackageManager: "winget",
		Supported:      true,
	}

	_, err := InstallCommand(profile)
	if err == nil {
		t.Fatal("InstallCommand(windows/winget) should return error")
	}
	if !strings.Contains(err.Error(), "POSIX") {
		t.Errorf("Windows error should explain POSIX requirement, got %q", err.Error())
	}
}
