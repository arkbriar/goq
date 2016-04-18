package main

import "os"

type Int int
type String string
type IntArray []int
type MapIntToString map[int]string
type IntChan chan int
type IntPtr *int
type File *os.File

type SomeInt (int)

type A struct {
	datai int
	dataf float32
}

type B interface {
	m(a int) error
}

func C() {

}
