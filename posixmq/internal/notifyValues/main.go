//go:build cgo

package main

import (
	"encoding/binary"
	"os"
)

/*
#include <signal.h>
*/
import "C"

func main() {
	var b []byte
	for _, i := range []int{C.SIGEV_NONE, C.SIGEV_SIGNAL} {
		b = binary.AppendVarint(b, int64(i))
	}

	if err := os.WriteFile("notifyValues.bin", b, os.ModePerm); err != nil {
		panic(err)
	}
}
