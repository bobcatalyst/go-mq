package sys

import (
	"fmt"
	"golang.org/x/sys/unix"
)

type Errnos map[unix.Errno]Value

func NewErrnos(errs []Value) Errnos {
	m := Errnos{}
	for _, e := range errs {
		en := e.Errno()
		if _, ok := m[en]; ok {
			panic(fmt.Errorf("duplicate errno %d", en))
		}
		m[en] = e
	}
	return m
}

// Get retrieves an error associated with a given unix.Errno.
// Returns nil if errno is 0, or the original errno if it is not mapped.
func (e Errnos) Get(errno unix.Errno) error {
	if err, ok := e[errno]; ok {
		return Errno{errno: err}
	} else if errno == 0 {
		// Syscall always returns an error, with a value of 0 meaning success.
		// Convert the syscall success indicator to a nil error for Go idiomatic handling.
		return nil
	}
	return errno
}

type Value interface {
	errnoValue
	errno
}

type errnoValue interface {
	error
	Errno() unix.Errno
}

type Errno struct {
	errno errnoValue
}

func (e Errno) Unwrap() error   { return e.errno }
func (e Errno) Temporary() bool { return e.errno.Errno().Temporary() }
func (e Errno) Timeout() bool   { return e.errno.Errno().Timeout() }
func (e Errno) Error() string   { return e.errno.Error() }

type errno interface {
	errno() unix.Errno
}

type Err[T errnoValue] struct{}

func (e Err[T]) errno() unix.Errno { return (*new(T)).Errno() }
func (e Err[T]) Unwrap() error     { return e.errno() }
