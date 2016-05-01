package pkgtest

type LinkNode struct {
	*FileAStruct
	prev *LinkNode
	next *LinkNode
}

func CreateLinkNode(d int) *LinkNode {
	return &LinkNode{FileAStruct: &FileAStruct{data: d}, prev: nil, next: nil}
}

func (this *LinkNode) FuncA() {

}

func (this *LinkNode) FuncB(x int) bool {
	return this.data == x
}

func RandomLinkedList() (root *LinkNode) {
	return nil
}
