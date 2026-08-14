package cpu

type HaltReason uint8

const (
	HaltNone HaltReason = iota
	HaltStopped
	HaltError

	// HaltExceptionStorm means the exception handler was itself faulting,
	// so execution was stopped to preserve the original cause.
	HaltExceptionStorm

	// HaltUnimplemented means a valid instruction was decoded that the
	// emulator does not implement yet.
	HaltUnimplemented

	// HaltBootROMReturn means the firmware returned to the address the
	// boot ROM would have supplied as its return address.
	HaltBootROMReturn

	// HaltWatchdogReset means the guest triggered a watchdog reset,
	// i.e. a system reboot.
	HaltWatchdogReset
)

func (h HaltReason) String() string {
	switch h {
	case HaltNone:
		return "none"
	case HaltStopped:
		return "stopped"
	case HaltError:
		return "error"
	case HaltExceptionStorm:
		return "exception storm"
	case HaltUnimplemented:
		return "unimplemented instruction"
	case HaltBootROMReturn:
		return "returned to boot ROM"
	case HaltWatchdogReset:
		return "watchdog reset"
	}
	return "unknown"
}
