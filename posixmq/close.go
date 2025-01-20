package posixmq

import (
	"github.com/bobcatalyst/go-mq/internal/sys"
	"golang.org/x/sys/unix"
)

var sysClose = sys.New(
	unix.SYS_CLOSE, 1,
	ErrBadFileDescriptor{},
)

// RawClose closes a message queue descriptor.
func RawClose(mqd int) error {
	return sysClose.Call(uintptr(mqd))
}
