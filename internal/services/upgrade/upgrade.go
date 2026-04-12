package upgrade

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	repoOwner = "Sadham-Hussian"
	repoName  = "recall"
	apiURL    = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
	userAgent = "recall-upgrade"
)

type Release struct {
	TagName string  `json:"tag_name"`
	HTMLURL string  `json:"html_url"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// LatestRelease fetches metadata for the most recent GitHub release.
func LatestRelease(ctx context.Context) (*Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned %s", resp.Status)
	}

	var r Release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}
	return &r, nil
}

// IsNewer compares two semver-ish strings (tolerates a leading "v" and
// build suffixes like "-3-gabc123-dirty" produced by `git describe`).
// A "dev" or empty current version always reports newer=true so that
// self-built binaries can still exercise the upgrade path.
func IsNewer(current, latest string) bool {
	if current == "" || current == "dev" {
		return true
	}
	cur := parseSemver(current)
	lat := parseSemver(latest)
	for i := 0; i < 3; i++ {
		if lat[i] > cur[i] {
			return true
		}
		if lat[i] < cur[i] {
			return false
		}
	}
	return false
}

func parseSemver(s string) [3]int {
	s = strings.TrimPrefix(s, "v")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.SplitN(s, ".", 3)
	var out [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		out[i], _ = strconv.Atoi(parts[i])
	}
	return out
}

// AssetName returns the archive name produced by goreleaser for this
// runtime's GOOS/GOARCH, e.g. "recall_darwin_arm64.tar.gz".
func AssetName() string {
	return fmt.Sprintf("recall_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
}

// Download fetches both the platform tarball and checksums.txt into temp
// files. Callers must Remove both returned paths.
func Download(ctx context.Context, release *Release) (tarballPath, checksumsPath string, err error) {
	want := AssetName()
	var tarballURL, checksumsURL string
	for _, a := range release.Assets {
		switch a.Name {
		case want:
			tarballURL = a.BrowserDownloadURL
		case "checksums.txt":
			checksumsURL = a.BrowserDownloadURL
		}
	}
	if tarballURL == "" {
		return "", "", fmt.Errorf("no asset %s in release %s", want, release.TagName)
	}
	if checksumsURL == "" {
		return "", "", fmt.Errorf("no checksums.txt in release %s", release.TagName)
	}

	tarballPath, err = downloadToTemp(ctx, tarballURL, want)
	if err != nil {
		return "", "", err
	}
	checksumsPath, err = downloadToTemp(ctx, checksumsURL, "checksums.txt")
	if err != nil {
		os.Remove(tarballPath)
		return "", "", err
	}
	return tarballPath, checksumsPath, nil
}

func downloadToTemp(ctx context.Context, url, name string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: %s", url, resp.Status)
	}

	f, err := os.CreateTemp("", "recall-upgrade-*-"+name)
	if err != nil {
		return "", err
	}
	path := f.Name()
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(path)
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}

// VerifyChecksum matches the SHA-256 of tarballPath against the entry
// in checksums.txt for the runtime's asset name.
func VerifyChecksum(tarballPath, checksumsPath string) error {
	want := AssetName()

	data, err := os.ReadFile(checksumsPath)
	if err != nil {
		return err
	}

	var expected string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if fields[1] == want {
			expected = fields[0]
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("no checksum entry for %s", want)
	}

	f, err := os.Open(tarballPath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if got != expected {
		return fmt.Errorf("checksum mismatch: want %s got %s", expected, got)
	}
	return nil
}

// Extract pulls the "recall" file out of the gzipped tarball into destDir
// and returns the full path to the extracted binary.
func Extract(tarballPath, destDir string) (string, error) {
	f, err := os.Open(tarballPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) != "recall" {
			continue
		}

		out := filepath.Join(destDir, "recall")
		w, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(w, tr); err != nil {
			w.Close()
			return "", err
		}
		if err := w.Close(); err != nil {
			return "", err
		}
		return out, nil
	}
	return "", fmt.Errorf("recall binary not found in archive")
}

// SwapBinary atomically replaces the running executable with newBinaryPath.
// Staging in the same directory guarantees os.Rename stays on one filesystem
// (required for atomicity). On Unix the old inode stays alive for the
// currently-running process, so self-replacement is safe.
func SwapBinary(newBinaryPath string) error {
	target, err := os.Executable()
	if err != nil {
		return err
	}
	target, err = filepath.EvalSymlinks(target)
	if err != nil {
		return err
	}

	if mgr, ok := IsManagedInstall(target); ok {
		switch mgr {
		case "homebrew":
			return fmt.Errorf("recall is managed by Homebrew; run `brew upgrade Sadham-Hussian/recall/recall` instead")
		default:
			return fmt.Errorf("recall is managed by %s; use that package manager to upgrade", mgr)
		}
	}

	stagedF, err := os.CreateTemp(filepath.Dir(target), ".recall.new.*")
	if err != nil {
		return fmt.Errorf("stage new binary in %s (re-run with sudo if not writable): %w", filepath.Dir(target), err)
	}
	staged := stagedF.Name()
	stagedF.Close()

	if err := copyFile(newBinaryPath, staged, 0755); err != nil {
		os.Remove(staged)
		return err
	}

	if err := os.Rename(staged, target); err != nil {
		os.Remove(staged)
		return fmt.Errorf("swap binary: %w", err)
	}
	return nil
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Chmod(dst, perm)
}

// IsManagedInstall reports true when the binary lives inside a path managed
// by a package manager whose integrity checks our in-place swap would break.
// Currently detects Homebrew (both Intel /usr/local/Cellar and Apple Silicon
// /opt/homebrew/Cellar layouts).
func IsManagedInstall(binaryPath string) (manager string, ok bool) {
	if strings.Contains(binaryPath, "/Cellar/") {
		return "homebrew", true
	}
	return "", false
}
