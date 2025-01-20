package sys

import (
	"fmt"
	"golang.org/x/sys/unix"
)

// syscall represents a low-level system call wrapper.
// It includes the syscall number, parameter count, and error mappings.
type syscall struct {
	trap   uintptr
	params int
	errs   Errnos
}

// New creates a new syscall instance.
// trap: The syscall number.
// params: The number of expected parameters.
// errnos: Optional error values mapped to specific error numbers.
func New(trap uintptr, params int, errnos ...Value) *syscall {
	return &syscall{
		trap:   trap,
		params: params,
		errs:   NewErrnos(errnos),
	}
}

// Call executes the system call and returns an error if one occurs.
func (s *syscall) Call(args ...uintptr) error {
	_, err := s.CallValue(args...)
	return err
}

// CallValue executes the system call and returns a single return value and an error.
func (s *syscall) CallValue(args ...uintptr) (int, error) {
	r, _, err := s.CallValues(args...)
	return r, err
}

// CallValues executes the system call and returns two return values and an error.
func (s *syscall) CallValues(args ...uintptr) (int, int, error) {
	var r1, r2 uintptr
	var en unix.Errno
	if len(args) != s.params {
		// Ensure the number of arguments matches the expected parameter count.
		return 0, 0, fmt.Errorf("got %d params, expected %d", len(args), s.params)
	} else if len(args) <= 3 {
		// Handle syscalls with up to 3 parameters using unix.Syscall.
		args = exactlyLen(args, 3)
		r1, r2, en = unix.Syscall(s.trap, args[0], args[1], args[2])
	} else {
		// Handle syscalls with up to 6 parameters using unix.Syscall6.
		args = exactlyLen(args, 6)
		r1, r2, en = unix.Syscall6(s.trap, args[0], args[1], args[2], args[3], args[4], args[5])
	}
	// Convert the errno into a Go error using the error mappings.
	return int(r1), int(r2), s.errs.Get(en)
}

func exactlyLen[T any](in []T, length int) []T {
	out := make([]T, length)
	copy(out, in)
	return out
}
