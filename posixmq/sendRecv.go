package posixmq

import (
	"github.com/bobcatalyst/go-mq/internal/deadline"
	"github.com/bobcatalyst/go-mq/internal/sys"
	"golang.org/x/sys/unix"
	"unsafe"
)

type ErrSendRecvTimeout struct {
	sys.Err[ErrSendRecvTimeout]
}

func (ErrSendRecvTimeout) Errno() unix.Errno { return unix.ETIMEDOUT }
func (ErrSendRecvTimeout) Error() string {
	return "the call timed out before a message could be transferred"
}

type ErrSendRecvInterrupted struct {
	sys.Err[ErrSendRecvInterrupted]
}

func (ErrSendRecvInterrupted) Errno() unix.Errno { return unix.EINTR }
func (ErrSendRecvInterrupted) Error() string {
	return "the call was interrupted by a signal handler"
}

type ErrSendRecvInvalidTimeout struct {
	sys.Err[ErrSendRecvInvalidTimeout]
}

func (ErrSendRecvInvalidTimeout) Errno() unix.Errno { return unix.EINVAL }
func (ErrSendRecvInvalidTimeout) Error() string {
	return "the call would have blocked and abs_timeout was invalid"
}

type ErrSendFullQueue struct {
	sys.Err[ErrSendFullQueue]
}

func (ErrSendFullQueue) Errno() unix.Errno { return unix.EAGAIN }
func (ErrSendFullQueue) Error() string {
	return "the queue was full and the O_NONBLOCK flag was set for the message queue"
}

type ErrRecvEmptyQueue struct {
	sys.Err[ErrRecvEmptyQueue]
}

func (ErrRecvEmptyQueue) Errno() unix.Errno { return unix.EAGAIN }
func (ErrRecvEmptyQueue) Error() string {
	return "the queue was empty and the O_NONBLOCK flag was set for the message queue"
}

type ErrSendInvalidFileDescriptor struct {
	sys.Err[ErrSendInvalidFileDescriptor]
}

func (ErrSendInvalidFileDescriptor) Errno() unix.Errno { return unix.EBADF }
func (ErrSendInvalidFileDescriptor) Error() string {
	return "the file descriptor was invalid or not opened for writing"
}

type ErrReceiveInvalidFileDescriptor struct {
	sys.Err[ErrReceiveInvalidFileDescriptor]
}

func (ErrReceiveInvalidFileDescriptor) Errno() unix.Errno { return unix.EBADF }
func (ErrReceiveInvalidFileDescriptor) Error() string {
	return "the file descriptor was invalid or not opened for reading"
}

type ErrSendInvalidMessageSize struct {
	sys.Err[ErrSendInvalidMessageSize]
}

func (ErrSendInvalidMessageSize) Errno() unix.Errno { return unix.EMSGSIZE }
func (ErrSendInvalidMessageSize) Error() string {
	return "msg_len was greater than the mq_msgsize attribute of the message queue"
}

type ErrRecvInvalidMessageSize struct {
	sys.Err[ErrRecvInvalidMessageSize]
}

func (ErrRecvInvalidMessageSize) Errno() unix.Errno { return unix.EMSGSIZE }
func (ErrRecvInvalidMessageSize) Error() string {
	return "msg_len was less than the mq_msgsize attribute of the message queue"
}

var (
	sysSend = sys.New(
		unix.SYS_MQ_TIMEDSEND, 5,
		ErrSendRecvTimeout{},
		ErrSendRecvInterrupted{},
		ErrSendRecvInvalidTimeout{},
		ErrSendFullQueue{},
		ErrSendInvalidFileDescriptor{},
		ErrSendInvalidMessageSize{},
	)
	sysRecv = sys.New(
		unix.SYS_MQ_TIMEDRECEIVE, 5,
		ErrSendRecvTimeout{},
		ErrSendRecvInterrupted{},
		ErrSendRecvInvalidTimeout{},
		ErrRecvEmptyQueue{},
		ErrReceiveInvalidFileDescriptor{},
		ErrRecvInvalidMessageSize{},
	)
)

// RawSendReceive sends or receives a message depending on the type of P.
//   - uint: Sends buf with the priority to mpd. On success (0, nil) is returned.
//   - *uint: Receives into buf from mqd and stores the priority in the provided pointer. The int returned is the size of the message.
//
// If dl does not return a Deadline, the operation will block unless [OpenNonBlocking] was specified.
func RawSendReceive[P uint | *uint](mqd int, dl deadline.Deadline, buf []byte, priority P) (int, error) {
	t, err := deadline.ToTimespec(dl)
	if err != nil {
		return 0, err
	}

	// Handle the priority type using a type switch to distinguish between send and receive operations.
	switch priority := any(priority).(type) {
	case uint:
		// Sending a message to the queue.
		return 0, sysSend.Call(
			uintptr(mqd),                     // mqdes
			uintptr(unsafe.Pointer(&buf[0])), // msg_ptr
			uintptr(len(buf)),                // msg_len
			uintptr(priority),                // msg_prio
			uintptr(unsafe.Pointer(t)),       // abs_timeout
		)
	case *uint:
		// Receiving a message from the queue.
		return sysRecv.CallValue(
			uintptr(mqd),                      // mqdes
			uintptr(unsafe.Pointer(&buf[0])),  // msg_ptr
			uintptr(len(buf)),                 // msg_len
			uintptr(unsafe.Pointer(priority)), // msg_prio
			uintptr(unsafe.Pointer(t)),        // abs_timeout
		)
	default:
		// This case should never be reached due to the type constraint on P.
		panic("unreachable")
	}
}
