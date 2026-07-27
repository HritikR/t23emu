package cpu

type HaltReason uint8

const (
	HaltNone HaltReason = iota
	HaltStopped
	HaltError
)
