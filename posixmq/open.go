package posixmq

import (
	"github.com/bobcatalyst/go-mq/internal/sys"
	"golang.org/x/sys/unix"
	"slices"
	"strings"
	"unsafe"
)

// OpenFlag represents the flags used for opening a message queue.
type OpenFlag int

var openFlagValueMap = map[OpenFlag]string{
	OpenWriteOnly:   "O_WRONLY",
	OpenReadWrite:   "O_RDWR",
	OpenCloseOnExec: "O_CLOEXEC",
	OpenCreate:      "O_CREAT",
	OpenExclusive:   "O_EXEC",
	OpenNonBlocking: "O_NONBLOCK",
}

// String identifies the set flags and combines them into a readable format.
// Only valid flags will be in the output.
func (f OpenFlag) String() string {
	flagStrings := make([]string, 0, 10)
	for value, name := range openFlagValueMap {
		if f&value == value {
			flagStrings = append(flagStrings, name)
		}
	}

	// Default to O_RDONLY if no specific read/write flag is set.
	if f&OpenWriteOnly != OpenWriteOnly && f&OpenReadWrite != OpenReadWrite {
		flagStrings = append(flagStrings, "O_RDONLY")
	}
	slices.Sort(flagStrings)
	return strings.Join(flagStrings, "|")
}

const (
	OpenReadOnly  OpenFlag = unix.O_RDONLY
	OpenWriteOnly OpenFlag = unix.O_WRONLY
	OpenReadWrite OpenFlag = unix.O_RDWR

	OpenCloseOnExec OpenFlag = unix.O_CLOEXEC
	OpenCreate      OpenFlag = unix.O_CREAT
	OpenExclusive   OpenFlag = unix.O_EXCL
	OpenNonBlocking OpenFlag = unix.O_NONBLOCK

	DefaultOpenFlags   = OpenReadOnly | OpenCloseOnExec
	DefaultCreateFlags = OpenWriteOnly | OpenCloseOnExec | OpenCreate | OpenExclusive
)

type ErrOpenBadAccess struct {
	sys.Err[ErrOpenBadAccess]
}

func (ErrOpenBadAccess) Errno() unix.Errno { return unix.EACCES }
func (ErrOpenBadAccess) Error() string {
	return "the queue exists, but the caller does not have permission to open it in the specified mode"
}

type ErrOpenExists struct {
	sys.Err[ErrOpenExists]
}

func (ErrOpenExists) Errno() unix.Errno { return unix.EEXIST }
func (ErrOpenExists) Error() string {
	return "both O_CREAT and O_EXCL were specified in oflag, but a queue with this name already exists"
}

type ErrOpenInvalid struct {
	sys.Err[ErrOpenInvalid]
}

func (ErrOpenInvalid) Errno() unix.Errno { return unix.EINVAL }
func (ErrOpenInvalid) Error() string {
	return "create flag was specified and attr->mq_maxmsg or attr->mq_msqsize were invalid"
}

type ErrOpenProcessLimitReached struct {
	sys.Err[ErrOpenProcessLimitReached]
}

func (ErrOpenProcessLimitReached) Errno() unix.Errno { return unix.EMFILE }
func (ErrOpenProcessLimitReached) Error() string {
	return "the per-process limit on the number of open file and message queue descriptors has been reached"
}

type ErrOpenSystemLimitReached struct {
	sys.Err[ErrOpenSystemLimitReached]
}

func (ErrOpenSystemLimitReached) Errno() unix.Errno { return unix.ENFILE }
func (ErrOpenSystemLimitReached) Error() string {
	return "the system-wide limit on the total number of open files and message queues has been reached"
}

type ErrOpenNoEntry struct {
	sys.Err[ErrOpenNoEntry]
}

func (ErrOpenNoEntry) Errno() unix.Errno { return unix.ENOENT }
func (ErrOpenNoEntry) Error() string {
	return "the O_CREAT flag was not specified in oflag, and no queue with this name exists"
}

type ErrOpenNoSpace struct {
	sys.Err[ErrOpenNoSpace]
}

func (ErrOpenNoSpace) Errno() unix.Errno { return unix.ENOSPC }
func (ErrOpenNoSpace) Error() string {
	return "insufficient space for the creation of a new message queue"
}

var sysOpen = sys.New(
	unix.SYS_MQ_OPEN, 4,
	ErrOpenBadAccess{},
	ErrOpenExists{},
	ErrOpenInvalid{},
	ErrOpenProcessLimitReached{},
	ErrNameTooLong{},
	ErrOpenSystemLimitReached{},
	ErrOpenNoEntry{},
	ErrNoMemory{},
	ErrOpenNoSpace{})

// RawOpen validates the name and opens a message queue.
// One of [OpenReadOnly], [OpenWriteOnly], or [OpenReadWrite] must be set.
// If opening with [OpenCreate] mode and attr must be set.
func RawOpen(name string, oflag OpenFlag, mode int, attr *Attributes) (int, error) {
	bn, err := namePtrFromString(name)
	if err != nil {
		return 0, err
	}
	return rawOpen(bn, oflag, mode, attr)
}

// rawOpen performs the low-level syscall to open a message queue.
func rawOpen(name *byte, oflag OpenFlag, mode int, attr *Attributes) (int, error) {
	return sysOpen.CallValue(
		uintptr(unsafe.Pointer(name)), // name
		uintptr(oflag),                // oflag
		uintptr(mode),                 // mode
		uintptr(unsafe.Pointer(attr)), // attr
	)
}
