package posixmq

import (
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
	"testing"
	"time"
)

const testNotifySig unix.Signal = unix.SIGUSR1

func TestRawNotify(t *testing.T) {
	name := randName()
	mq, err := RawOpen(name, OpenCreate|OpenReadWrite, 0644, &Attributes{
		MaxQueueSize:   1,
		MaxMessageSize: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer RawUnlink(name)
	defer RawClose(mq)

	c := make(chan os.Signal, 1)
	signal.Notify(c, testNotifySig)
	errc := make(chan error, 1)

	if err := RawNotify(mq, &Notify{Signo: testNotifySig}); err != nil {
		t.Fatal(err)
	}

	go func() {
		if _, err := RawSendReceive[uint](mq, t, []byte{1}, 1); err != nil {
			errc <- err
		}
	}()

	for range 5 {
		select {
		case err := <-errc:
			t.Fatal(err)
		case sig := <-c:
			if sig != testNotifySig {
				t.Fatalf("expected %[1]q(%[1]d), got %[2]q(%[2]d)", testNotifySig, sig)
			} else {
				t.Logf("notified %[1]q(%[1]d)", sig)
			}
			return
		case <-time.Tick(time.Second * 5):
			t.Log("notify timeout")
		}
	}
	t.Fail()
}
