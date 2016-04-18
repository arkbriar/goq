package codelib

func __assert(condition bool, msg string) {
	if !condition {
		panic(msg)
	}
}
