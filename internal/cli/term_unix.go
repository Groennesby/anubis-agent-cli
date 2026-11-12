//go:build !windows

package cli

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"unsafe"
)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func ioctlTermSize() (int, int) {
	ws := &winsize{}
	ret, _, _ := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))
	if int(ret) == 0 && ws.Col > 0 && ws.Row > 0 {
		return int(ws.Col), int(ws.Row)
	}
	return 0, 0
}

func RegisterResizeSignal(ch chan os.Signal) {
	signal.Notify(ch, syscall.SIGWINCH)
}

// ReExecSelf replaces the currently running process in memory with the newly updated binary on disk.
func ReExecSelf() {
	home, _ := os.UserHomeDir()
	userBin := filepath.Join(home, ".anubis", "bin", "anubis")
	if st, err := os.Stat(userBin); err == nil && !st.IsDir() {
		_ = os.Setenv("ANUBIS_DELEGATED", "1")
		_ = syscall.Exec(userBin, os.Args, os.Environ())
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if realPath, err := filepath.EvalSymlinks(exe); err == nil && realPath != "" {
		exe = realPath
	}
	_ = syscall.Exec(exe, os.Args, os.Environ())
}

// DelegateToNewerUserBinary checks at startup if ~/.anubis/bin/anubis was updated by /update and is newer than /usr/bin/anubis.
func DelegateToNewerUserBinary() {
	if os.Getenv("ANUBIS_DELEGATED") == "1" {
		return
	}
	home, _ := os.UserHomeDir()
	userBin := filepath.Join(home, ".anubis", "bin", "anubis")
	uStat, err := os.Stat(userBin)
	if err != nil || uStat.IsDir() {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if realPath, lErr := filepath.EvalSymlinks(exe); lErr == nil && realPath != "" {
		exe = realPath
	}
	if exe == userBin {
		return
	}
	if eStat, sErr := os.Stat(exe); sErr == nil && uStat.ModTime().After(eStat.ModTime()) {
		_ = os.Setenv("ANUBIS_DELEGATED", "1")
		_ = syscall.Exec(userBin, os.Args, os.Environ())
	}
}
