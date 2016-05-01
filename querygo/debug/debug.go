// Copyright 2016 ArkBriar. All rights reserved.
package debug

import (
	"log"
	"os"
)

type DebugLog bool

var dbgLog = log.New(os.Stdout, "[DEBUG] ", log.Ltime)

func (d DebugLog) Printf(format string, args ...interface{}) {
	if d {
		dbgLog.Printf(format, args...)
	}
}

func (d DebugLog) Println(args ...interface{}) {
	if d {
		dbgLog.Println(args...)
	}
}
