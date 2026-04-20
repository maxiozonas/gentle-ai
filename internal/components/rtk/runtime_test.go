package rtk

import (
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gentleman-programming/gentle-ai/internal/system"
)

func TestAvailable_FoundViaLookPath(t *testing.T) {
	origLookPath := cmdLookPath
	origStat := osStat
	origHomeDir := osUserHomeDir
	t.Cleanup(func() {
		cmdLookPath = origLookPath
		osStat = origStat
		osUserHomeDir = origHomeDir
	})

	cmdLookPath = func(name string) (string, error) {
		return "/usr/local/bin/rtk", nil
	}

	profile := system.PlatformProfile{OS: "darwin", PackageManager: "brew"}
	if !Available(profile) {
		t.Error("Available() = false when rtk is on PATH")
	}
}

func TestAvailable_FoundViaHomeLocalBin(t *testing.T) {
	origLookPath := cmdLookPath
	origStat := osStat
	origHomeDir := osUserHomeDir
	t.Cleanup(func() {
		cmdLookPath = origLookPath
		osStat = origStat
		osUserHomeDir = origHomeDir
	})

	cmdLookPath = func(name string) (string, error) {
		return "", exec.ErrNotFound
	}
	osStat = func(name string) (os.FileInfo, error) {
		if name == "/home/test/.local/bin/rtk" {
			return fakeFileInfo{}, nil
		}
		return nil, errors.New("not found")
	}
	osUserHomeDir = func() (string, error) {
		return "/home/test", nil
	}

	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}
	if !Available(profile) {
		t.Error("Available() = false when rtk is in ~/.local/bin/")
	}
}

func TestAvailable_FoundViaHomebrewPrefix(t *testing.T) {
	origLookPath := cmdLookPath
	origStat := osStat
	origHomeDir := osUserHomeDir
	t.Cleanup(func() {
		cmdLookPath = origLookPath
		osStat = origStat
		osUserHomeDir = origHomeDir
	})

	cmdLookPath = func(name string) (string, error) {
		return "", exec.ErrNotFound
	}
	osStat = func(name string) (os.FileInfo, error) {
		if name == "/opt/homebrew/bin/rtk" {
			return fakeFileInfo{}, nil
		}
		return nil, errors.New("not found")
	}
	osUserHomeDir = func() (string, error) {
		return "/Users/test", nil
	}

	profile := system.PlatformProfile{OS: "darwin", PackageManager: "brew"}
	if !Available(profile) {
		t.Error("Available() = false when rtk is in /opt/homebrew/bin/")
	}
}

func TestAvailable_NotFound(t *testing.T) {
	origLookPath := cmdLookPath
	origStat := osStat
	origHomeDir := osUserHomeDir
	t.Cleanup(func() {
		cmdLookPath = origLookPath
		osStat = origStat
		osUserHomeDir = origHomeDir
	})

	cmdLookPath = func(name string) (string, error) {
		return "", exec.ErrNotFound
	}
	osStat = func(name string) (os.FileInfo, error) {
		return nil, errors.New("not found")
	}
	osUserHomeDir = func() (string, error) {
		return "/home/test", nil
	}

	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}
	if Available(profile) {
		t.Error("Available() = true when rtk is not installed")
	}
}

func TestAvailable_HomeDirError(t *testing.T) {
	origLookPath := cmdLookPath
	origStat := osStat
	origHomeDir := osUserHomeDir
	t.Cleanup(func() {
		cmdLookPath = origLookPath
		osStat = origStat
		osUserHomeDir = origHomeDir
	})

	cmdLookPath = func(name string) (string, error) {
		return "", exec.ErrNotFound
	}
	osUserHomeDir = func() (string, error) {
		return "", errors.New("no home")
	}

	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}
	if Available(profile) {
		t.Error("Available() should return false when home dir cannot be resolved")
	}
}

func TestVerifyInstallation_Success(t *testing.T) {
	orig := rtkVersionOutput
	t.Cleanup(func() { rtkVersionOutput = orig })

	rtkVersionOutput = func() ([]byte, error) {
		return []byte("rtk 1.2.3\n"), nil
	}

	if err := VerifyInstallation(); err != nil {
		t.Errorf("VerifyInstallation() error = %v", err)
	}
}

func TestVerifyInstallation_VersionWithVPrefix(t *testing.T) {
	orig := rtkVersionOutput
	t.Cleanup(func() { rtkVersionOutput = orig })

	rtkVersionOutput = func() ([]byte, error) {
		return []byte("v2.0.1\n"), nil
	}

	if err := VerifyInstallation(); err != nil {
		t.Errorf("VerifyInstallation() error = %v", err)
	}
}

func TestVerifyInstallation_ExecError(t *testing.T) {
	orig := rtkVersionOutput
	t.Cleanup(func() { rtkVersionOutput = orig })

	rtkVersionOutput = func() ([]byte, error) {
		return nil, errors.New("binary not found")
	}

	if err := VerifyInstallation(); err == nil {
		t.Error("VerifyInstallation() should fail when exec fails")
	}
}

func TestVerifyInstallation_EmptyOutput(t *testing.T) {
	orig := rtkVersionOutput
	t.Cleanup(func() { rtkVersionOutput = orig })

	rtkVersionOutput = func() ([]byte, error) {
		return []byte(""), nil
	}

	if err := VerifyInstallation(); err == nil {
		t.Error("VerifyInstallation() should fail on empty output")
	}
}

func TestVerifyInstallation_UnexpectedOutput(t *testing.T) {
	orig := rtkVersionOutput
	t.Cleanup(func() { rtkVersionOutput = orig })

	rtkVersionOutput = func() ([]byte, error) {
		return []byte("not a version string"), nil
	}

	if err := VerifyInstallation(); err == nil {
		t.Error("VerifyInstallation() should fail on non-semver output")
	}
}

func TestVerifyVersion_Success(t *testing.T) {
	orig := rtkVersionOutput
	t.Cleanup(func() { rtkVersionOutput = orig })

	rtkVersionOutput = func() ([]byte, error) {
		return []byte("rtk 3.14.159\n"), nil
	}

	version, err := VerifyVersion()
	if err != nil {
		t.Fatalf("VerifyVersion() error = %v", err)
	}
	if version != "3.14.159" {
		t.Errorf("VerifyVersion() = %q, want %q", version, "3.14.159")
	}
}

func TestVerifyVersion_ExecError(t *testing.T) {
	orig := rtkVersionOutput
	t.Cleanup(func() { rtkVersionOutput = orig })

	rtkVersionOutput = func() ([]byte, error) {
		return nil, errors.New("exit status 1")
	}

	if _, err := VerifyVersion(); err == nil {
		t.Error("VerifyVersion() should fail when exec fails")
	}
}

// fakeFileInfo is a minimal os.FileInfo implementation for tests.
type fakeFileInfo struct{}

func (fakeFileInfo) Name() string       { return "rtk" }
func (fakeFileInfo) Size() int64        { return 0 }
func (fakeFileInfo) Mode() os.FileMode  { return 0o755 }
func (fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (fakeFileInfo) IsDir() bool        { return false }
func (fakeFileInfo) Sys() interface{}   { return nil }
