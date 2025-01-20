package posixmq

import (
	"github.com/bobcatalyst/go-mq/internal/sys"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"
)

const (
	AttributeNonBlocking = AttributeFlag(OpenNonBlocking)
	AttributeBlocking    = AttributeFlag(0)
)

type AttributeFlag OpenFlag

// String converts the AttributeFlag to a human-readable string.
func (af AttributeFlag) String() string {
	if af&AttributeNonBlocking == AttributeNonBlocking {
		return "O_NONBLOCK"
	}
	return ""
}

type Attributes struct {
	// Flags can only be 0 or [OpenNonBlocking].
	Flags AttributeFlag `json:"mq_flags"`
	// MaxQueueSize is the max size of the queue.
	MaxQueueSize int `json:"mq_maxmsg"`
	// MaxMessageSize is the max size of a message.
	MaxMessageSize int `json:"mq_msgsize"`
	// NumCurrMessages is the number of messages in the queue.
	NumCurrMessages int `json:"mq_curmsgs"`
}

type ErrGetSetAttrInvalidFlags struct {
	sys.Err[ErrGetSetAttrInvalidFlags]
}

func (ErrGetSetAttrInvalidFlags) Errno() unix.Errno { return unix.EINVAL }
func (ErrGetSetAttrInvalidFlags) Error() string {
	return "newattr->mq_flags contained set bits other than O_NONBLOCK"
}

var sysGetSetAttr = sys.New(
	unix.SYS_MQ_GETSETATTR, 3,
	ErrBadFileDescriptor{},
	ErrGetSetAttrInvalidFlags{},
)

// RawGetSetAttributes retrieves or modifies the attributes of a message queue.
// - If newAttr is nil, it retrieves the current attributes of the queue.
// - If newAttr is provided, it updates the blocking flag and returns the attributes before the change.
func RawGetSetAttributes(mqd int, newAttr *Attributes) (attr Attributes, err error) {
	err = sysGetSetAttr.Call(
		uintptr(mqd),                     // mqdes
		uintptr(unsafe.Pointer(newAttr)), // newattr
		uintptr(unsafe.Pointer(&attr)),   // oldattr
	)
	return
}

func DefaultMessageSize() (int, error) { return sizeFromFile("msqsize_default") }
func MaxMessageSize() (int, error)     { return sizeFromFile("msqsize_max") }
func DefaultQueueSize() (int, error)   { return sizeFromFile("msq_default") }
func MaxQueueSize() (int, error)       { return sizeFromFile("msq_max") }
func MaxQueues() (int, error)          { return sizeFromFile("queues_max") }

func sizeFromFile(name string) (int, error) {
	b, err := os.ReadFile(filepath.Join("/proc/sys/fs/mqueue", name))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(b)))
}
