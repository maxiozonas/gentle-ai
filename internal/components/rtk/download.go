package rtk

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gentleman-programming/gentle-ai/internal/system"
)

const (
	rtkOwner      = "rtk-ai"
	rtkRepo       = "rtk"
	rtkBinaryName = "rtk"
)

var (
	rtkHTTPClient    = &http.Client{Timeout: 5 * time.Minute}
	rtkGitHubBaseURL = "https://github.com"
	rtkInstallDirFn  = rtkInstallDir
)

// DownloadLatestBinary fetches the latest rtk release from GitHub and
// installs it to ~/.local/bin (or $RTK_INSTALL_DIR if set, matching upstream).
// It returns the full path to the installed binary.
//
// This is the non-brew installation path for Linux (apt/pacman/dnf).
// On macOS with Homebrew, brew install rtk is used instead.
// Windows is not supported — RTK hooks require a POSIX shell.
func DownloadLatestBinary(profile system.PlatformProfile) (string, error) {
	if profile.OS == "windows" {
		return "", fmt.Errorf("rtk does not support Windows native — install gentle-ai inside WSL")
	}

	version, err := fetchLatestRTKVersion()
	if err != nil {
		return "", fmt.Errorf("fetch latest rtk version: %w", err)
	}

	target, err := rtkTarget(profile.OS, runtime.GOARCH)
	if err != nil {
		return "", err
	}

	assetURL := rtkAssetURL(rtkGitHubBaseURL, version, target)

	installDir := rtkInstallDirFn()
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return "", fmt.Errorf("create rtk install dir %q: %w", installDir, err)
	}

	outPath := filepath.Join(installDir, rtkBinaryName)

	if err := downloadAndExtractRTKTarGz(assetURL, rtkBinaryName, outPath); err != nil {
		return "", fmt.Errorf("download rtk tar.gz: %w", err)
	}

	return outPath, nil
}

// fetchLatestRTKVersion queries the GitHub Releases API for the latest rtk
// release and returns the version tag (e.g. "v0.37.2"). The tag is used
// verbatim in the asset URL (rtk tags include the leading "v").
func fetchLatestRTKVersion() (string, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/%s/releases/latest",
		rtkAPIBaseURL(), rtkOwner, rtkRepo)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token := rtkGithubToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := rtkHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decode release JSON: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("empty tag_name in GitHub release response")
	}

	return release.TagName, nil
}

func rtkGithubToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	return os.Getenv("GH_TOKEN")
}

// rtkTarget maps Go OS/arch to the Rust target triple used by rtk release assets.
// Mirrors the detection logic in rtk's official install.sh.
func rtkTarget(goos, goarch string) (string, error) {
	switch goos {
	case "linux":
		switch goarch {
		case "amd64", "386":
			return "x86_64-unknown-linux-musl", nil
		case "arm64", "arm":
			return "aarch64-unknown-linux-gnu", nil
		}
	case "darwin":
		switch goarch {
		case "amd64":
			return "x86_64-apple-darwin", nil
		case "arm64":
			return "aarch64-apple-darwin", nil
		}
	}
	return "", fmt.Errorf("unsupported platform for rtk: os=%q arch=%q", goos, goarch)
}

// rtkAPIBaseURL returns the GitHub API base URL. In tests a mock server
// handles both API and download under the same URL, so we derive the API
// base from rtkGitHubBaseURL when it points to localhost.
func rtkAPIBaseURL() string {
	base := rtkGitHubBaseURL
	if strings.Contains(base, "127.0.0.1") || strings.Contains(base, "localhost") {
		return base
	}
	return "https://api.github.com"
}

// rtkAssetURL constructs the download URL. rtk tags include the leading "v"
// in the URL path (e.g. releases/download/v0.37.2/rtk-...tar.gz), so we use
// the tag verbatim.
func rtkAssetURL(baseURL, tag, target string) string {
	return fmt.Sprintf("%s/%s/%s/releases/download/%s/%s-%s.tar.gz",
		baseURL, rtkOwner, rtkRepo, tag, rtkBinaryName, target)
}

// rtkInstallDir returns the directory where the rtk binary should be installed.
// Matches rtk's official install.sh default: $RTK_INSTALL_DIR or ~/.local/bin.
func rtkInstallDir() string {
	if env := os.Getenv("RTK_INSTALL_DIR"); env != "" {
		return env
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/usr/local/bin"
	}
	return filepath.Join(home, ".local", "bin")
}

func downloadAndExtractRTKTarGz(url, binaryName, outPath string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := rtkHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	return extractRTKBinaryFromTarGz(resp.Body, binaryName, outPath)
}

// extractRTKBinaryFromTarGz reads a .tar.gz stream and extracts the first
// regular file whose base name matches binaryName, writing it to outPath
// with executable permissions via atomic rename (avoids ETXTBSY on Linux).
func extractRTKBinaryFromTarGz(r io.Reader, binaryName, outPath string) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("open gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		if filepath.Base(hdr.Name) == binaryName &&
			(hdr.Typeflag == tar.TypeReg || hdr.Typeflag == tar.TypeRegA) {
			return writeRTKExecutable(tr, outPath)
		}
	}

	return fmt.Errorf("binary %q not found in archive", binaryName)
}

// writeRTKExecutable writes the content from r to outPath with executable
// permissions using an atomic rename. Mirrors engram's writeExecutable to
// avoid ETXTBSY errors when replacing a running binary.
func writeRTKExecutable(r io.Reader, outPath string) error {
	dir := filepath.Dir(outPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".rtk-upgrade-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	defer func() {
		if tmpPath != "" {
			os.Remove(tmpPath)
		}
	}()

	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", tmpPath, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}

	if err := os.Rename(tmpPath, outPath); err != nil {
		return fmt.Errorf("rename %s -> %s: %w", tmpPath, outPath, err)
	}

	tmpPath = ""
	return nil
}
