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
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if realPath, err := filepath.EvalSymlinks(exe); err == nil && realPath != "" {
		exe = realPath
	}
	_ = syscall.Exec(exe, os.Args, os.Environ())
}
