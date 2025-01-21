package posixmq

import (
    "errors"
    "fmt"
    "github.com/bobcatalyst/go-mq/internal/deadline"
    "golang.org/x/sys/unix"
    "sync"
)

// MQ allows for structured usage of POSIX message queues.
type MQ struct {
    name  string      // Name of the message queue.
    bname *byte       // Byte pointer representation of the name.
    attr  *Attributes // Queue attributes, nil unless OpenCreate was passed as an oflag.
    mode  int         // Mode used when creating the queue.
    oflag OpenFlag    // Flags used to open the queue.

    mqd    int          // Message queue descripto
    buf    []byte       // Internal buffer for receiving messages.
    close  func() error // Function to close the queue once.
    unlink func() error // Function to unlink the queue once.
}

// MQOption represents options that can be applied when creating or opening a message queue.
type MQOption interface {
    applyOption(*MQ)
}

type (
    optionOflag      OpenFlag
    optionCreateArgs struct {
        mode           int
        maxMessageSize int
        maxQueueSize   int
    }
)

// OptionCreateArgs supplies arguments for creating a queue with [OpenCreate].
func OptionCreateArgs(mode, maxMessageSize, maxQueueSize int) MQOption {
    return &optionCreateArgs{
        mode:           mode,
        maxMessageSize: maxMessageSize,
        maxQueueSize:   maxQueueSize,
    }
}

// OptionOflag sets the oflag parameter for opening the queue.
func OptionOflag(oflag OpenFlag) MQOption { return optionOflag(oflag) }

func (opt optionOflag) applyOption(mq *MQ) { mq.oflag = OpenFlag(opt) }
func (opt *optionCreateArgs) applyOption(mq *MQ) {
    mq.mode = opt.mode
    mq.attr = &Attributes{
        MaxQueueSize:   opt.maxQueueSize,
        MaxMessageSize: opt.maxMessageSize,
    }
}

// New opens a message queue.
func New(name string, opts ...MQOption) (mq *MQ, err error) {
    mq = &MQ{name: name}
    if mq.bname, err = namePtrFromString(name); err != nil {
        return nil, err
    }

    if err := mq.applyAll(opts); err != nil {
        return nil, err
    }
    if err := mq.open(); err != nil {
        return nil, err
    }
    return mq, nil
}

func (mq *MQ) applyAll(opts []MQOption) error {
    for _, opt := range opts {
        opt.applyOption(mq)
    }

    if mq.attr != nil && mq.oflag&OpenCreate != OpenCreate {
        mq.oflag |= OpenCreate
    }

    if mq.oflag&OpenCreate == OpenCreate {
        if mq.mode == 0 {
            mq.mode = 0644
        }
        if mq.attr == nil {
            defQueueSize, qsErr := DefaultQueueSize()
            defMsqSize, msErr := DefaultMessageSize()
            if err := errors.Join(qsErr, msErr); err != nil {
                return err
            }

            mq.attr = &Attributes{
                MaxQueueSize:   defQueueSize,
                MaxMessageSize: defMsqSize,
            }
        }
    } else {
        mq.mode = 0
        mq.attr = nil
    }
    return nil
}

// open initializes the queue and sets up close and unlink operations.
func (mq *MQ) open() (err error) {
    if mq.mqd, err = rawOpen(mq.bname, mq.oflag, mq.mode, mq.attr); err != nil {
        return err
    }

    mq.unlink = func() error { return rawUnlink(mq.bname) }
    mq.close = sync.OnceValue(func() error { return RawClose(mq.mqd) })
    return nil
}

// Send sends a message to the queue.
func (mq *MQ) Send(dl deadline.Deadline, data []byte, priority uint) error {
    _, err := RawSendReceive(mq.mqd, dl, data, priority)
    return err
}

// Receive retrieves a message from the queue.
// The returned data is invalid after the next call to Receive.
func (mq *MQ) Receive(dl deadline.Deadline) (data []byte, priority uint, _ error) {
    if len(mq.buf) == 0 {
        // Receive buffer has not been initialized yet.
        // The only option changeable on an open message queue is blocking, so this only needs to be done once.
        attr, err := mq.GetAttr()
        if err != nil {
            return nil, 0, fmt.Errorf("failed to get message buffer size from attributes: %w", err)
        } else if attr.MaxMessageSize <= 0 {
            return nil, 0, fmt.Errorf("invalid MaxMessageSize of %d", attr.MaxMessageSize)
        }
        mq.buf = make([]byte, attr.MaxMessageSize)
    }
    size, err := RawSendReceive(mq.mqd, dl, mq.buf, &priority)
    return mq.buf[:size], priority, err
}

// Name returns the name of the queue.
func (mq *MQ) Name() string {
    return mq.name
}

// Mode returns the mode used to create the queue. Returns 0 if [OpenCreate] was not used.
func (mq *MQ) Mode() int {
    return mq.mode
}

// Oflag returns the flags used to open the queue.
func (mq *MQ) Oflag() OpenFlag {
    return mq.oflag
}

// Mqd returns the message queue descriptor for advanced usage.
func (mq *MQ) Mqd() int {
    return mq.mqd
}

// Close closes the message queue.
func (mq *MQ) Close() error {
    return mq.close()
}

// Unlink closes and unlinks the queue. The system will free it once all processes close it.
func (mq *MQ) Unlink() error {
    err := mq.Close()
    return errors.Join(err, mq.unlink())
}

// GetAttr gets the message queue's attributes.
func (mq *MQ) GetAttr() (oldValue Attributes, _ error) {
    return RawGetSetAttributes(mq.mqd, nil)
}

// SetBlocking sets or clears the blocking flag on the message queue.
func (mq *MQ) SetBlocking(blocking bool) (Attributes, error) {
    var attr Attributes
    if !blocking {
        attr.Flags = AttributeNonBlocking
    }
    return RawGetSetAttributes(mq.mqd, &attr)
}

// Notify sets up notifications for the queue using a signal.
func (mq *MQ) Notify(sig unix.Signal) error {
    return RawNotify(mq.mqd, &Notify{
        Notify: NotifySignal,
        Signo:  sig,
    })
}

// ClearNotify clears any registered notifications.
func (mq *MQ) ClearNotify() error {
    return RawNotify(mq.mqd, nil)
}
