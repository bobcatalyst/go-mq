package posixmq

import (
	"github.com/bobcatalyst/go-mq/internal/sys"
	"golang.org/x/sys/unix"
)

type ErrBadFileDescriptor struct {
	sys.Err[ErrBadFileDescriptor]
}

func (e ErrBadFileDescriptor) Errno() unix.Errno { return unix.EBADF }
func (e ErrBadFileDescriptor) Error() string {
	return "the message queue descriptor specified is invalid"
}

type ErrNameTooLong struct {
	sys.Err[ErrNameTooLong]
}

func (err ErrNameTooLong) Errno() unix.Errno { return unix.ENAMETOOLONG }
func (err ErrNameTooLong) Error() string {
	return "name was too long"
}

type ErrNoMemory struct {
	sys.Err[ErrNoMemory]
}

func (err ErrNoMemory) Errno() unix.Errno { return unix.ENOMEM }
func (err ErrNoMemory) Error() string {
	return "insufficient memory"
}
