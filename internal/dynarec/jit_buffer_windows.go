//go:build windows

package dynarec

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procVirtualAlloc = kernel32.NewProc("VirtualAlloc")
	procVirtualFree  = kernel32.NewProc("VirtualFree")
)

const (
	memCommit            = 0x1000
	memReserve           = 0x2000
	memRelease           = 0x8000
	pageExecuteReadWrite = 0x40
)

// NewJITBuffer allocates an executable RWX memory page using Windows VirtualAlloc.
func NewJITBuffer(size int) (*JITBuffer, error) {
	if size <= 0 {
		size = 4096
	}

	addr, _, err := procVirtualAlloc.Call(
		0,
		uintptr(size),
		memCommit|memReserve,
		pageExecuteReadWrite,
	)
	if addr == 0 {
		return nil, fmt.Errorf("jit: VirtualAlloc failed: %w", err)
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(addr)), size)
	return &JITBuffer{
		data: data,
		ptr:  addr,
	}, nil
}

func (b *JITBuffer) freeOS() error {
	if b.ptr != 0 {
		r, _, err := procVirtualFree.Call(b.ptr, 0, memRelease)
		if r == 0 {
			return fmt.Errorf("jit: VirtualFree failed: %w", err)
		}
		b.data = nil
		b.ptr = 0
	}
	return nil
}
