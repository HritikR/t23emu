//go:build darwin || linux || freebsd || netbsd || openbsd

package dynarec

import (
	"fmt"
	"syscall"
)

// NewJITBuffer allocates an executable memory page using mmap.
func NewJITBuffer(size int) (*JITBuffer, error) {
	if size <= 0 {
		size = 4096
	}

	pageSize := syscall.Getpagesize()
	size = (size + pageSize - 1) & ^(pageSize - 1)

	prot := syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC
	flags := syscall.MAP_ANON | syscall.MAP_PRIVATE

	data, err := syscall.Mmap(-1, 0, size, prot, flags)
	if err != nil {
		return nil, fmt.Errorf("jit: mmap failed: %w", err)
	}

	return &JITBuffer{data: data}, nil
}

func (b *JITBuffer) freeOS() error {
	if b.data != nil {
		err := syscall.Munmap(b.data)
		b.data = nil
		return err
	}
	return nil
}
