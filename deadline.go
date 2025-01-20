package go_mq

import "github.com/bobcatalyst/go-mq/internal/deadline"

// Deadline is used to abstract deadline management in message queue operations.
type Deadline = deadline.Deadline

// NoDeadline is a predefined deadline that indicates no specific time constraints.
// Operations using NoDeadline will block unless other conditions are met.
var NoDeadline = deadline.NoDeadline{}

// TimeDeadline represents a deadline with a specific time constraint.
// It allows setting a precise timeout for operations.
type TimeDeadline = deadline.TimeDeadline
