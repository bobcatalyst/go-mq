package go_mq

import "github.com/bobcatalyst/go-mq/internal/sys"

// Errno provides an abstraction for error codes returned by system calls.
// It allows for introspection without requiring knowledge of the specific error type.
type Errno = sys.Errno
