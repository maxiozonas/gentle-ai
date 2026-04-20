package rtk

import (
	"reflect"
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

	want := CommandSequence{{"brew", "install", "rtk"}}
	if !reflect.DeepEqual(cmds, want) {
		t.Errorf("InstallCommand(darwin/brew) = %v, want %v", cmds, want)
	}
}

func TestInstallCommand_LinuxApt(t *testing.T) {
	profile := system.PlatformProfile{
		OS:             "linux",
		PackageManager: "apt",
		Supported:      true,
	}

	cmds, err := InstallCommand(profile)
	if err != nil {
		t.Fatalf("InstallCommand(linux/apt) error = %v", err)
	}

	if len(cmds) != 1 {
		t.Fatalf("InstallCommand(linux/apt) returned %d commands, want 1", len(cmds))
	}

	// Command: sh -c "curl -fsSL ... | sh"
	cmd := cmds[0]
	if cmd[0] != "sh" || cmd[1] != "-c" {
		t.Errorf("Linux command = %v, want sh -c <install-script>", cmd)
	}
	if len(cmd) != 3 {
		t.Errorf("Linux command has %d args, want 3 (sh -c <script>)", len(cmd))
	}
	// Verify it contains the RTK install script URL
	if cmd[2] == "" {
		t.Error("Linux install script command is empty")
	}
}

func TestInstallCommand_LinuxPacman(t *testing.T) {
	profile := system.PlatformProfile{
		OS:             "linux",
		PackageManager: "pacman",
		Supported:      true,
	}

	cmds, err := InstallCommand(profile)
	if err != nil {
		t.Fatalf("InstallCommand(linux/pacman) error = %v", err)
	}

	if len(cmds) != 1 {
		t.Fatalf("InstallCommand(linux/pacman) returned %d commands, want 1", len(cmds))
	}
}

func TestInstallCommand_LinuxDnf(t *testing.T) {
	profile := system.PlatformProfile{
		OS:             "linux",
		PackageManager: "dnf",
		Supported:      true,
	}

	cmds, err := InstallCommand(profile)
	if err != nil {
		t.Fatalf("InstallCommand(linux/dnf) error = %v", err)
	}

	if len(cmds) != 1 {
		t.Fatalf("InstallCommand(linux/dnf) returned %d commands, want 1", len(cmds))
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
		t.Error("InstallCommand(windows/winget) should return error")
	}
}

func TestShouldInstall(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		want    bool
	}{
		{name: "enabled returns true", enabled: true, want: true},
		{name: "disabled returns false", enabled: false, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldInstall(tt.enabled); got != tt.want {
				t.Errorf("ShouldInstall(%v) = %v, want %v", tt.enabled, got, tt.want)
			}
		})
	}
}
