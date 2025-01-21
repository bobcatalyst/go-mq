package posixmq

import (
    "testing"
)

func TestRawSendReceive(t *testing.T) {
    type args struct {
        buf      []byte
        priority any
    }

    type testCase struct {
        name    string
        args    args
        want    int
        wantErr bool
    }

    for _, tt := range []testCase{
        // TODO: Add test cases.
    } {
        t.Run(tt.name, func(t *testing.T) {
            var err error
            var got int
            switch p := tt.args.priority.(type) {
            case uint:
                got, err = RawSendReceive(tt.args.mqd, t, tt.args.buf, p)
            case *uint:
                got, err = RawSendReceive(tt.args.mqd, t, tt.args.buf, p)
            default:
                t.FailNow()
            }
            if (err != nil) != tt.wantErr {
                t.Errorf("RawSendReceive() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("RawSendReceive() got = %v, want %v", got, tt.want)
            }
        })
    }
}
