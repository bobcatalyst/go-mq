package posixmq

import (
	"fmt"
	"github.com/bobcatalyst/go-mq/internal/sys"
	"golang.org/x/sys/unix"
	"math/rand"
	"strings"
)

type ErrNameContainedMultipleSlash struct {
	sys.Err[ErrNameContainedMultipleSlash]
}

func (ErrNameContainedMultipleSlash) Errno() unix.Errno { return unix.EACCES }

func (ErrNameContainedMultipleSlash) Error() string {
	return "name contained more than one slash"
}

type ErrNameEmpty struct {
	sys.Err[ErrNameEmpty]
}

func (ErrNameEmpty) Errno() unix.Errno { return unix.ENOENT }

func (ErrNameEmpty) Error() string {
	return "name was just '/' followed by no other characters"
}

type ErrNameInvalid struct {
	sys.Err[ErrNameInvalid]
}

func (ErrNameInvalid) Errno() unix.Errno { return unix.EINVAL }

func (ErrNameInvalid) Error() string {
	return "name doesn't follow the correct format"
}

// namePtrFromString converts a queue name to a pointer for system calls.
// Validates the name before conversion.
// Name must not contain any \x00.
func namePtrFromString(name string) (*byte, error) {
	name, err := ValidateName(name)
	if err != nil {
		return nil, err
	}
	// Helper to convert from name to pointer to the beginning of the string bytes.
	return unix.BytePtrFromString(name)
}

// ValidateName checks if a name is valid and returns it without the initial slash.
// The name must start with '/', be non-empty, and contain no other '/' characters.
func ValidateName(name string) (string, error) {
	if len(name) == 0 {
		// Check if the name is empty.
		return "", fmt.Errorf("%w, name must not be empty", ErrNameInvalid{})
	} else if name[0] != '/' {
		// Check if the name doesn't start with a '/'.
		return "", fmt.Errorf("%w, name must start with '/'", ErrNameInvalid{})
	} else if len(name) == 1 {
		// Check if the name is just '/'.
		return "", ErrNameEmpty{}
	} else if strings.Contains(name[1:], "/") {
		// Check if the name contains any '/' characters.
		return "", ErrNameContainedMultipleSlash{}
	}
	return name[1:], nil
}

func randName() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVQXYZ1234567890"
	var sb strings.Builder
	sb.WriteRune('/')
	for range 30 {
		sb.WriteByte(chars[rand.Intn(len(chars))])
	}
	sb.WriteString(".tmp")
	return sb.String()
}
