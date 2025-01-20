package posixmq

import (
	"github.com/bobcatalyst/go-mq/internal/sys"
	"golang.org/x/sys/unix"
	"unsafe"
)

type ErrUnlinkNoPermission struct {
	sys.Err[ErrUnlinkNoPermission]
}

func (ErrUnlinkNoPermission) Errno() unix.Errno { return unix.EACCES }
func (ErrUnlinkNoPermission) Error() string {
	return "the caller does not have permission to unlink this message queue"
}

type ErrUnlinkNoMessageQueue struct {
	sys.Err[ErrUnlinkNoMessageQueue]
}

func (ErrUnlinkNoMessageQueue) Errno() unix.Errno { return unix.ENOENT }
func (ErrUnlinkNoMessageQueue) Error() string {
	return "there is no message queue with the given name"
}

// RawUnlink removes the message queue with the specified name from the system.
// The system will free the queue once all instances are closed.
func RawUnlink(name string) error {
	bn, err := namePtrFromString(name)
	if err != nil {
		return err
	}
	return rawUnlink(bn)
}

var sysUnlink = sys.New(
	unix.SYS_MQ_UNLINK, 1,
	ErrUnlinkNoPermission{},
	ErrNameTooLong{},
	ErrUnlinkNoMessageQueue{},
)

// rawUnlink performs the low-level unlink syscall using the provided byte pointer.
func rawUnlink(bn *byte) error {
	return sysUnlink.Call(
		uintptr(unsafe.Pointer(bn)), // name
	)
}
