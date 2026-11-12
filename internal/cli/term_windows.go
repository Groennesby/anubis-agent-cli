//go:build windows

package cli

import "os"

func ioctlTermSize() (int, int) {
	return 120, 36
}

func RegisterResizeSignal(ch chan os.Signal) {
	// Windows console resize handled via environment / console API fallback
}
