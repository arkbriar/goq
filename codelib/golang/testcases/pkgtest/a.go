package pkgtest

type FileAInterface interface {
	FuncA()
	FuncB(x int) bool
}

type FileAStruct struct {
	data int
}

func (this *FileAStruct) GetData() int {
	return this.data
}

func FileAFunc() int {
	return 42
}
