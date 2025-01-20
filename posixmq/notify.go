package posixmq

import (
	_ "embed"
	"encoding/binary"
	"github.com/bobcatalyst/go-mq/internal/sys"
	"golang.org/x/sys/unix"
	"unsafe"
)

//go:generate go run internal/notifyValues/main.go

var (
	NotifyNone   int // Registers a handler but sends no signal
	NotifySignal int
	//go:embed notifyValues.bin
	notifyValues []byte
)

func init() {
	b := notifyValues
	for _, p := range []*int{&NotifyNone, &NotifySignal} {
		nn, n := binary.Uvarint(b)
		b = b[n:]
		*p = int(nn)
	}
}

// Notify defines the structure for queue notification settings.
type Notify struct {
	Notify int         `json:"sigev_notify"` // Notification type (e.g., NotifyNone, NotifySignal).
	Signo  unix.Signal `json:"sigev_signo"`  // Signal number to use for signal-based notifications.
}

type ErrNotifyBusy struct {
	sys.Err[ErrNotifyBusy]
}

func (ErrNotifyBusy) Errno() unix.Errno { return unix.EBUSY }
func (ErrNotifyBusy) Error() string {
	return "another process has already registered to receive notification for this message queue"
}

type ErrNotifyInvalid struct {
	sys.Err[ErrNotifyInvalid]
}

func (ErrNotifyInvalid) Errno() unix.Errno { return unix.EINVAL }
func (ErrNotifyInvalid) Error() string {
	return "sevp->sigev_notify is not one of the permitted values or sevp->sigev_notify is SIGEV_SIGNAL and sevp->sigev_signo is not a valid signal number"
}

var sysNotify = sys.New(
	unix.SYS_MQ_NOTIFY, 2,
	ErrBadFileDescriptor{},
	ErrNoMemory{},
	ErrNotifyBusy{},
	ErrNotifyInvalid{},
)

// RawNotify sets up a notification mechanism for a message queue.
func RawNotify(mq int, notify *Notify) error {
	return sysNotify.Call(uintptr(mq), uintptr(unsafe.Pointer(notify)))
}
