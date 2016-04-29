// Copyright 2016 ArkBriar. All rights reserved.
package querygo

import (
	"os"
	"strings"
)

var (
	GOPATH = strings.Split(os.Getenv("GOPATH"), ":")
	GOROOT = strings.Split(os.Getenv("GOROOT"), ":")
)
