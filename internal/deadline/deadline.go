package deadline

import (
	"golang.org/x/sys/unix"
	"time"
)

// Deadline is an interface representing something that may provide a deadline.
// Examples include [context.Context] or [testing.T].
type Deadline interface {
	Deadline() (time.Time, bool)
}

// NoDeadline is a type that implements the Deadline interface but always indicates no deadline.
type NoDeadline struct{}

func (NoDeadline) Deadline() (_ time.Time, _ bool) { return }

// TimeDeadline contains a simple [time.Time] value and indicates a deadline if the time is not a zero value.
type TimeDeadline time.Time

func (td TimeDeadline) Deadline() (time.Time, bool) {
	return time.Time(td), !time.Time(td).IsZero()
}

// ToTimespec converts a Deadline to a [unix.Timespec].
// If the Deadline has a valid, non-zero deadline, it converts the deadline time to a Timespec.
// Returns nil if no deadline is set or if the deadline is zero.
func ToTimespec(dl Deadline) (*unix.Timespec, error) {
	if t, ok := dl.Deadline(); ok && !t.IsZero() {
		ts, err := unix.TimeToTimespec(t)
		return &ts, err
	}
	return nil, nil
}
