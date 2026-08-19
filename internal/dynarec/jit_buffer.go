package dynarec

// JITBuffer manages executable RWX memory pages.
type JITBuffer struct {
	data []byte
	ptr  uintptr
}

// Bytes returns the raw executable memory slice.
func (b *JITBuffer) Bytes() []byte {
	return b.data
}

// Free releases the allocated memory.
func (b *JITBuffer) Free() error {
	return b.freeOS()
}
