package posixmq

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func TestMQ_GetAttr(t *testing.T) {
	ms := rand.Intn(50) + 1
	qs := rand.Intn(10) + 1
	for _, test := range []struct {
		name  string
		fn    func(*testing.T, *MQ)
		oflag OpenFlag
	}{
		{
			name: "get attrs",
			fn: func(t *testing.T, mq *MQ) {
				attr, err := mq.GetAttr()
				if err != nil {
					t.Fatal(err)
				}
				assertAttrs(t, &attr, &Attributes{
					MaxQueueSize:   qs,
					MaxMessageSize: ms,
				})
				logAttrs(t, &attr)
			},
		},
		{
			name:  "non blocking flag",
			oflag: OpenNonBlocking | DefaultCreateFlags,
			fn: func(t *testing.T, mq *MQ) {
				attr, err := mq.GetAttr()
				if err != nil {
					t.Fatal(err)
				}
				assertAttrs(t, &attr, &Attributes{
					MaxQueueSize:   qs,
					MaxMessageSize: ms,
					Flags:          AttributeNonBlocking,
				})
				logAttrs(t, &attr)
			},
		},
		{
			name:  "with messages",
			oflag: OpenWriteOnly,
			fn: func(t *testing.T, mq *MQ) {
				ns := rand.Intn(qs) + 1
				for i := range ns {
					if err := mq.Send(t, []byte{byte(i)}, uint(i)); err != nil {
						t.Fatal(err)
					}
				}

				attr, err := mq.GetAttr()
				if err != nil {
					t.Fatal(err)
				}
				assertAttrs(t, &attr, &Attributes{
					MaxQueueSize:    qs,
					MaxMessageSize:  ms,
					NumCurrMessages: ns,
				})
				logAttrs(t, &attr)
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			opts := []MQOption{OptionCreateArgs(0644, ms, qs)}
			if test.oflag != 0 {
				opts = append(opts, OptionOflag(test.oflag))
			}
			mq, err := New(randName(), opts...)
			if err != nil {
				t.Fatal(err)
				return
			}
			defer mq.Unlink()
			test.fn(t, mq)
		})
	}
}

func logAttrs(t *testing.T, attr *Attributes) {
	t.Helper()
	var vs []string
	rt := reflect.TypeOf(attr).Elem()
	rv := reflect.ValueOf(attr).Elem()
	for i := range rt.NumField() {
		ft := rt.Field(i)
		fv := rv.Field(i)
		vs = append(vs, fmt.Sprintf("%s(%v)", ft.Name, fv.Interface()))
	}
	t.Log(strings.Join(vs, " "))
}

func assertAttrs(t *testing.T, attr, expected *Attributes) {
	t.Helper()
	rt := reflect.TypeOf(attr).Elem()
	rv := reflect.ValueOf(attr).Elem()
	rve := reflect.ValueOf(expected).Elem()
	for i := range rt.NumField() {
		ft := rt.Field(i)
		fv := rv.Field(i)
		fve := rve.Field(i)

		if !fv.Equal(fve) {
			t.Fatalf("expected %s to be %v, got %v", ft.Name, fve.Interface(), fv.Interface())
		}
	}
}

func TestMQ_SetBlocking(t *testing.T) {
	ms := rand.Intn(50) + 1
	qs := rand.Intn(10) + 1
	for _, test := range []struct {
		name  string
		oflag OpenFlag
		fn    func(*testing.T, *MQ)
	}{
		{
			name: "set blocking",
			fn: func(t *testing.T, mq *MQ) {
				attr, err := mq.SetBlocking(true)
				if err != nil {
					t.Fatal(err)
				}
				assertAttrs(t, &attr, &Attributes{
					MaxQueueSize:   qs,
					MaxMessageSize: ms,
				})
				logAttrs(t, &attr)
			},
		},
		{
			name: "set non blocking",
			fn: func(t *testing.T, mq *MQ) {
				attr, err := mq.SetBlocking(false)
				if err != nil {
					t.Fatal(err)
				}
				assertAttrs(t, &attr, &Attributes{
					MaxQueueSize:   qs,
					MaxMessageSize: ms,
				})

				attr, err = mq.GetAttr()
				if err != nil {
					t.Fatal(err)
				}
				assertAttrs(t, &attr, &Attributes{
					Flags:          AttributeNonBlocking,
					MaxQueueSize:   qs,
					MaxMessageSize: ms,
				})
				logAttrs(t, &attr)
			},
		},
		{
			name:  "set unset non blocking",
			oflag: OpenWriteOnly,
			fn: func(t *testing.T, mq *MQ) {
				attr, err := mq.SetBlocking(false)
				if err != nil {
					t.Fatal(err)
				}
				assertAttrs(t, &attr, &Attributes{
					MaxQueueSize:   qs,
					MaxMessageSize: ms,
				})

				attr, err = mq.GetAttr()
				if err != nil {
					t.Fatal(err)
				}
				assertAttrs(t, &attr, &Attributes{
					Flags:          AttributeNonBlocking,
					MaxQueueSize:   qs,
					MaxMessageSize: ms,
				})
				logAttrs(t, &attr)

				attr, err = mq.SetBlocking(true)
				if err != nil {
					t.Fatal(err)
				}
				assertAttrs(t, &attr, &Attributes{
					Flags:          AttributeNonBlocking,
					MaxQueueSize:   qs,
					MaxMessageSize: ms,
				})

				if err := mq.Send(t, []byte{1}, 0); err != nil {
					t.Fatal(err)
				}

				attr, err = mq.GetAttr()
				if err != nil {
					t.Fatal(err)
				}
				assertAttrs(t, &attr, &Attributes{
					MaxQueueSize:    qs,
					MaxMessageSize:  ms,
					NumCurrMessages: 1,
				})
				logAttrs(t, &attr)
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			opts := []MQOption{OptionCreateArgs(0644, ms, qs)}
			if test.oflag != 0 {
				opts = append(opts, OptionOflag(test.oflag))
			}
			mq, err := New(randName(), opts...)
			if err != nil {
				t.Fatal(err)
				return
			}
			defer mq.Unlink()
			test.fn(t, mq)
		})
	}
}
