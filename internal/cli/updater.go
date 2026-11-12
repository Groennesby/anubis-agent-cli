package cli

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// SelfUpdate updates the Anubis Agent CLI & GUI binaries in-place without requiring uninstall/reinstall.
// Works seamlessly on both macOS (/opt/homebrew/bin) and Linux/Ubuntu Server (/usr/bin, /usr/local/bin).
func SelfUpdate(t Theme) (string, error) {
	var logs []string
	localSrc := "/Users/macm1pro/Documents/Anubis Agent/anubis-agent-cli"
	goBin := "/opt/homebrew/bin/go"
	if _, err := os.Stat(goBin); err != nil {
		if p, lookErr := exec.LookPath("go"); lookErr == nil {
			goBin = p
		}
	}

	exePath, _ := os.Executable()
	if realPath, err := filepath.EvalSymlinks(exePath); err == nil && realPath != "" {
		exePath = realPath
	}

	targetPaths := []string{
		"/opt/homebrew/bin/annubis",
		"/opt/homebrew/bin/anubis",
		"/opt/homebrew/Cellar/annubis/1.0.0/bin/annubis",
		"/opt/homebrew/Cellar/annubis/1.0.0/bin/anubis",
		"/usr/bin/anubis",
		"/usr/bin/annubis",
		"/usr/local/bin/anubis",
		"/usr/local/bin/annubis",
	}
	if exePath != "" {
		targetPaths = append(targetPaths, exePath)
	}

	if st, err := os.Stat(localSrc); err == nil && st.IsDir() && goBin != "" {
		coreSrc := "/Users/macm1pro/Documents/Anubis Agent/anubis-agent-core"
		_ = exec.Command("git", "-C", coreSrc, "pull", "--rebase", "origin", "main").Run()
		_ = exec.Command("git", "-C", localSrc, "pull", "--rebase", "origin", "main").Run()

		tmpBin := filepath.Join(os.TempDir(), fmt.Sprintf("anubis-update-%d", time.Now().UnixNano()))
		cmd := exec.Command(goBin, "build", "-ldflags=-s -w", "-o", tmpBin, ".")
		cmd.Dir = localSrc
		if out, buildErr := cmd.CombinedOutput(); buildErr == nil {
			defer os.Remove(tmpBin)
			updatedCount := installBinaryToTargets(tmpBin, targetPaths)
			_ = copyExecutable(tmpBin, filepath.Join(localSrc, "bin", "anubis"))
			_ = copyExecutable(tmpBin, filepath.Join(localSrc, "bin", "annubis"))
			logs = append(logs, fmt.Sprintf("⬆️ [/update] Gecompileerd vanaf nieuwste broncode & %d binaire paden in-place geüpdatet!", updatedCount))
			return strings.Join(logs, "\n"), nil
		} else {
			logs = append(logs, fmt.Sprintf("Lokale build overgeslagen (%s), schakel over naar GitHub Release download...", strings.TrimSpace(string(out))))
		}
	}

	// Download latest release tarball directly from GitHub Releases (works on Ubuntu Server, Debian, Kali, macOS)
	arch := runtime.GOARCH
	goos := runtime.GOOS
	if goos == "windows" {
		return "⬆️ [/update] Gebruik op Windows: scoop update anubis of irm https://raw.githubusercontent.com/Groennesby/Anubis-Agent/main/install.ps1 | iex", nil
	}

	tarURL := fmt.Sprintf("https://github.com/Groennesby/Anubis-Agent/releases/download/v1.0.0/anubis-%s-%s.tar.gz?t=%d", goos, arch, time.Now().Unix())
	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Get(tarURL)
	if err != nil {
		return "", fmt.Errorf("kon update niet downloaden van %s: %w", tarURL, err)
	}
	defer resp.Body.Close()

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	tmpBin := filepath.Join(os.TempDir(), fmt.Sprintf("anubis-dl-%d", time.Now().UnixNano()))
	found := false
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Name == "anubis" || hdr.Name == "./anubis" {
			f, err := os.OpenFile(tmpBin, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
			if err != nil {
				return "", err
			}
			_, _ = io.Copy(f, tr)
			_ = f.Close()
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("geen anubis binary gevonden in release archief")
	}
	defer os.Remove(tmpBin)

	updatedCount := installBinaryToTargets(tmpBin, targetPaths)
	return fmt.Sprintf("⬆️ [/update] Nieuwste versie gedownload van GitHub Releases & %d binaire paden in-place geüpdatet!", updatedCount), nil
}

func installBinaryToTargets(tmpBin string, targetPaths []string) int {
	updatedCount := 0
	seen := map[string]bool{}
	for _, tp := range targetPaths {
		if seen[tp] {
			continue
		}
		seen[tp] = true
		dir := filepath.Dir(tp)
		if dir == "" {
			continue
		}
		if _, dErr := os.Stat(dir); dErr != nil {
			continue
		}
		// Only overwrite existing binary paths or standard bin paths on the current OS
		if _, fErr := os.Stat(tp); fErr != nil && !strings.HasPrefix(tp, "/opt/homebrew/bin") && !strings.HasPrefix(tp, "/usr/bin") {
			continue
		}
		_ = os.Remove(tp)
		if copyErr := copyExecutable(tmpBin, tp); copyErr == nil {
			updatedCount++
			continue
		}
		// On Linux/Ubuntu where /usr/bin/anubis is owned by root, fall back to sudo cp -f
		if runtime.GOOS == "linux" {
			if err := exec.Command("sudo", "-n", "cp", "-f", tmpBin, tp).Run(); err == nil {
				_ = exec.Command("sudo", "-n", "chmod", "755", tp).Run()
				updatedCount++
			} else if err := exec.Command("sudo", "cp", "-f", tmpBin, tp).Run(); err == nil {
				_ = exec.Command("sudo", "chmod", "755", tp).Run()
				updatedCount++
			}
		}
	}
	return updatedCount
}

func copyExecutable(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return os.Chmod(dst, 0o755)
}
