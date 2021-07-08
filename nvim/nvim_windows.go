// +build windows

package nvim

import (
	"syscall"
)

// BinaryName is the name of default nvim binary name.
const BinaryName = "nvim.exe"

func init() {
	embedProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
