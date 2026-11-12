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

// SelfUpdate updates the Anubis Agent CLI & GUI binaries in-place (100% standardized on 'anubis' with 1 'n').
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

	home, _ := os.UserHomeDir()
	userAnubisBin := filepath.Join(home, ".anubis", "bin", "anubis")
	userLocalBin := filepath.Join(home, ".local", "bin", "anubis")
	_ = os.MkdirAll(filepath.Dir(userAnubisBin), 0o755)
	_ = os.MkdirAll(filepath.Dir(userLocalBin), 0o755)

	targetPaths := []string{
		userAnubisBin,
		userLocalBin,
		"/opt/homebrew/bin/anubis",
		"/home/linuxbrew/.linuxbrew/bin/anubis",
		"/usr/bin/anubis",
		"/usr/local/bin/anubis",
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
			_ = copyExecutableAtomic(tmpBin, filepath.Join(localSrc, "bin", "anubis"))
			logs = append(logs, fmt.Sprintf("⬆️ [/update] Gecompileerd vanaf nieuwste broncode & %d 'anubis' paden in-place geüpdatet!", updatedCount))
			return strings.Join(logs, "\n"), nil
		} else {
			logs = append(logs, fmt.Sprintf("Lokale build overgeslagen (%s), schakel over naar GitHub Release download...", strings.TrimSpace(string(out))))
		}
	}

	arch := runtime.GOARCH
	goos := runtime.GOOS
	if goos == "windows" {
		return "⬆️ [/update] Gebruik op Windows: scoop update anubis", nil
	}

	tarURL := fmt.Sprintf("https://github.com/Groennesby/Anubis-Agent/releases/latest/download/anubis-%s-%s.tar.gz?t=%d", goos, arch, time.Now().Unix())
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
	return fmt.Sprintf("⬆️ [/update] Nieuwste versie gedownload van GitHub Releases & %d 'anubis' paden in-place geüpdatet!", updatedCount), nil
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
		if copyErr := copyExecutableAtomic(tmpBin, tp); copyErr == nil {
			updatedCount++
			continue
		}
		if runtime.GOOS == "linux" {
			if err := exec.Command("sudo", "-n", "cp", "-f", tmpBin, tp).Run(); err == nil {
				_ = exec.Command("sudo", "-n", "chmod", "755", tp).Run()
				updatedCount++
			}
		}
	}
	return updatedCount
}

func copyExecutableAtomic(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmpDst := fmt.Sprintf("%s.new.%d", dst, time.Now().UnixNano())
	out, err := os.OpenFile(tmpDst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(tmpDst)
		return err
	}
	_ = out.Close()
	_ = os.Chmod(tmpDst, 0o755)
	if err := os.Rename(tmpDst, dst); err != nil {
		_ = os.Remove(tmpDst)
		return err
	}
	return nil
}
