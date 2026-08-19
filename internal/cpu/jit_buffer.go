package cpu

import (
	"fmt"
	"syscall"
)

// JITBuffer manages executable RWX memory pages.
type JITBuffer struct {
	data []byte
}

// NewJITBuffer allocates an executable memory page using mmap.
func NewJITBuffer(size int) (*JITBuffer, error) {
	if size <= 0 {
		size = 4096
	}

	// Align to page size
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

// Bytes returns the raw executable memory slice.
func (b *JITBuffer) Bytes() []byte {
	return b.data
}

// Free releases the allocated memory.
func (b *JITBuffer) Free() error {
	if b.data != nil {
		err := syscall.Munmap(b.data)
		b.data = nil
		return err
	}
	return nil
}
