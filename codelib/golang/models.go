// Copyright 2016 ArkBriar. All rights reserved.
package golang

import (
	"errors"
	"fmt"
)

func _IsUpper(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

type nsMethods interface {
	AddType(__type *GoType) error
	AddFunc(__func *GoFunc) error
	GetTypes() []*GoType
	GetFuncs() []*GoFunc
	GetType(name string) *GoType
	GetFunc(name string) *GoFunc
	GetStructs() []*GoStruct
	GetInterfaces() []*GoInterface
}

type goNamespace struct {
	nsMethods

	Types map[string]*GoType // types
	Funcs map[string]*GoFunc // functions
}

func __CreateGoNameSpace() *goNamespace {
	gn := new(goNamespace)
	gn.Types = make(map[string]*GoType)
	gn.Funcs = make(map[string]*GoFunc)

	return gn
}

func (this *goNamespace) AddType(__type *GoType) error {
	if _, ok := this.Types[__type.Name()]; ok {
		return errors.New("Type " + __type.Name() + " already exists!")
	}

	this.Types[__type.Name()] = __type

	return nil
}

func (this *goNamespace) AddFunc(__func *GoFunc) error {
	if _, ok := this.Funcs[__func.Name]; ok {
		return errors.New("Function " + __func.Name + " already exists")
	}

	this.Funcs[__func.Name] = __func

	return nil
}

func (this *goNamespace) GetTypes() []*GoType {
	var types []*GoType = make([]*GoType, 0, len(this.Types))

	for _, v := range this.Types {
		types = append(types, v)
	}

	return types
}

func (this *goNamespace) GetStructs() []*GoStruct {
	var structs []*GoStruct = make([]*GoStruct, 0, 16)

	for _, v := range this.Types {
		if v.Kind == Stt {
			structs = append(structs, v.Type.(*GoStruct))
		}
	}

	return structs
}

func (this *goNamespace) GetInterfaces() []*GoInterface {
	var interfaces []*GoInterface = make([]*GoInterface, 0, 16)

	for _, v := range this.Types {
		if v.Kind == Itf {
			interfaces = append(interfaces, v.Type.(*GoInterface))
		}
	}

	return interfaces
}

func (this *goNamespace) GetFuncs() []*GoFunc {
	var funcs []*GoFunc = make([]*GoFunc, 0, len(this.Funcs))

	for _, v := range this.Funcs {
		funcs = append(funcs, v)
	}

	return funcs
}

func (this *goNamespace) GetType(name string) *GoType {
	if name[0] == '*' {
		name = name[1:]
	}
	return this.Types[name]
}

func (this *goNamespace) GetFunc(name string) *GoFunc {
	return this.Funcs[name]
}

// if the first rune of symbol is uppercase, then it's public; else private
func IsPublic(symbol string) bool {
	return _IsUpper(symbol[0])
}

type TypeKind int

const (
	Stt = 1 << iota // struct
	Itf             // interface
	Als             // alias
	Bti             // builtin
)

type GoType struct {
	Kind TypeKind
	Type interface{} // should be *GoStruct, *GoInterface, *GoBuiltin, *GoAlias
}

func (this *GoType) Name() string {
	var name string
	switch this.Kind {
	case Stt:
		name = this.Type.(*GoStruct).Name
	case Itf:
		name = this.Type.(*GoInterface).Name
	case Als:
		name = this.Type.(*GoAlias).Name
	case Bti:
		name = this.Type.(*GoBuiltin).Name
	default:
		panic("golang/models ## GoType.Name: should not reach here")
	}

	return name
}

func (this *GoType) String() string {
	return this.Type.(fmt.Stringer).String()
}

func CreateGoTypeOfStruct(__type *GoStruct) *GoType {
	return &GoType{
		Kind: Stt,
		Type: __type,
	}
}

func CreateGoTypeOfInterface(__type *GoInterface) *GoType {
	return &GoType{
		Kind: Itf,
		Type: __type,
	}
}

func CreateGoTypeOfAlias(__type *GoAlias) *GoType {
	return &GoType{
		Kind: Als,
		Type: __type,
	}
}

func CreateGoTypeOfBuiltin(__type *GoBuiltin) *GoType {
	return &GoType{
		Kind: Bti,
		Type: __type,
	}
}

var (
	// Builtin types & interfaces(e.g. type error interface { Error() string })
	Bool       = &GoType{Kind: Bti, Type: &GoBuiltin{"bool"}}
	Byte       = &GoType{Kind: Bti, Type: &GoBuiltin{"byte"}}
	Complex    = &GoType{Kind: Bti, Type: &GoBuiltin{"complex"}}
	Complex64  = &GoType{Kind: Bti, Type: &GoBuiltin{"complex64"}}
	Complex128 = &GoType{Kind: Bti, Type: &GoBuiltin{"complex128"}}
	Error      = &GoType{Kind: Bti, Type: &GoBuiltin{"error"}}
	Float32    = &GoType{Kind: Bti, Type: &GoBuiltin{"float32"}}
	Float64    = &GoType{Kind: Bti, Type: &GoBuiltin{"float64"}}
	Int        = &GoType{Kind: Bti, Type: &GoBuiltin{"int"}}
	Int16      = &GoType{Kind: Bti, Type: &GoBuiltin{"int16"}}
	Int32      = &GoType{Kind: Bti, Type: &GoBuiltin{"int32"}}
	Int64      = &GoType{Kind: Bti, Type: &GoBuiltin{"int64"}}
	Int8       = &GoType{Kind: Bti, Type: &GoBuiltin{"int8"}}
	Rune       = &GoType{Kind: Bti, Type: &GoBuiltin{"rune"}}
	String     = &GoType{Kind: Bti, Type: &GoBuiltin{"string"}}
	Uint       = &GoType{Kind: Bti, Type: &GoBuiltin{"uint"}}
	Uint16     = &GoType{Kind: Bti, Type: &GoBuiltin{"uint16"}}
	Uint32     = &GoType{Kind: Bti, Type: &GoBuiltin{"uint32"}}
	Uint64     = &GoType{Kind: Bti, Type: &GoBuiltin{"uint64"}}
	Uint8      = &GoType{Kind: Bti, Type: &GoBuiltin{"uint8"}}
	Uintptr    = &GoType{Kind: Bti, Type: &GoBuiltin{"uintptr"}}
)

type (
	GoPackage struct {
		Name         string             // package name
		RelativePath string             // relative path in this project
		Files        map[string]*GoFile // source files in this package
		GlobalNs     *goNamespace       // global namespace (all types & functions in this package)

		upper *GoPackage // upper package
	}

	GoFile struct {
		Name    string       // file name
		Ns      *goNamespace // local namespace (types & functions)
		Package string       // package name
		Imports []string     // imports
	}

	GoVar struct {
		Name string // var name
		Type string // var type
	}

	GoStruct struct {
		Name        string                  // struct name
		Vars        map[string]*GoVar       // variables in struct
		Methods     map[string]*GoMethod    // methods
		Interfaces  map[string]*GoInterface // interfaces
		Extends     map[string]*GoStruct    // extends
		__Anonymous []string                // anonymous, embedded structs or interfaces
	}

	GoInterface struct {
		Name        string                  // interface name
		Methods     map[string]*GoMethod    // methods
		Extends     map[string]*GoInterface // extend interfaces
		__Anonymous []string                // anonymous, embedded interfaces
	}

	GoAlias struct {
		Name       string                  // alias name
		Type       string                  // original type
		Methods    map[string]*GoMethod    // methods
		Interfaces map[string]*GoInterface // interfaces
	}

	GoBuiltin struct {
		Name string
	}

	GoFunc struct {
		Name string   // function name
		Args []*GoVar // function args
		Rets []*GoVar // function rets
	}

	GoMethod struct {
		GoFunc // extends function
		// who receives this method
		// Recv *GoStruct/GoAlias
	}
)

func (this *GoPackage) String() string {
	ret := "[PACKAGE]" + this.Name + ": \n"

	for _, file := range this.Files {
		ret += file.String() + "\n"
	}

	// delete the last '\n'
	if len(this.Files) != 0 {
		return ret[0 : len(ret)-1]
	}

	return ret
}

func (this *GoFile) String() string {
	return "[FILE]" + this.Name + " ## Types: " + fmt.Sprint(this.Ns.Types) + " | Funcs: " + fmt.Sprint(this.Ns.Funcs)
}

func CreateGoPackage(name string, relativePath string) *GoPackage {
	gp := &GoPackage{
		Name:         name,
		RelativePath: relativePath,
		Files:        make(map[string]*GoFile),
		GlobalNs:     __CreateGoNameSpace(),
		upper:        nil,
	}

	return gp
}

func (this *GoPackage) N() string {
	return this.Name
}

func CreateGoFile(name string) *GoFile {
	gf := &GoFile{
		Name:    name,
		Ns:      __CreateGoNameSpace(),
		Imports: make([]string, 0, 8),
	}

	return gf
}

func (this *GoFile) N() string {
	return this.Name
}

func CreateGoStruct(name string) *GoStruct {
	gs := &GoStruct{
		Name:        name,
		Vars:        make(map[string]*GoVar),
		Methods:     make(map[string]*GoMethod),
		Interfaces:  make(map[string]*GoInterface),
		Extends:     make(map[string]*GoStruct),
		__Anonymous: make([]string, 0, 2),
	}

	return gs
}

func (this *GoStruct) N() string {
	return this.Name
}

func (this *GoStruct) IsPublic() bool {
	return _IsUpper(this.Name[0])
}

func CreateGoInterface(name string) *GoInterface {
	gi := &GoInterface{
		Name:        name,
		Methods:     make(map[string]*GoMethod),
		Extends:     make(map[string]*GoInterface),
		__Anonymous: make([]string, 0, 2),
	}

	return gi
}

func (this *GoInterface) N() string {
	return this.Name
}

func (this *GoInterface) IsPublic() bool {
	return _IsUpper(this.Name[0])
}

func CreateGoAlias(name string, __type string) *GoAlias {
	ga := &GoAlias{
		Name:       name,
		Type:       __type,
		Methods:    make(map[string]*GoMethod),
		Interfaces: make(map[string]*GoInterface),
	}

	return ga
}

func (this *GoAlias) N() string {
	return this.Name
}

func (this *GoAlias) IsPublic() bool {
	return _IsUpper(this.Name[0])
}

func CreateGoFunc(name string) *GoFunc {
	gf := &GoFunc{
		Name: name,
		Args: make([]*GoVar, 0, 4),
		Rets: make([]*GoVar, 0, 2),
	}

	return gf
}

func (this *GoFunc) N() string {
	return this.Name
}

func (this *GoFunc) IsPublic() bool {
	return _IsUpper(this.Name[0])
}

func CreateGoMethod(name string) *GoMethod {
	return &GoMethod{GoFunc: *CreateGoFunc(name)}
}

func (this *GoStruct) AddMethod(method *GoMethod) {
	this.Methods[method.Name] = method
}

func (this *GoInterface) AddMethod(method *GoMethod) {
	this.Methods[method.Name] = method
}

func (this *GoAlias) AddMethod(method *GoMethod) {
	this.Methods[method.Name] = method
}

func (this *GoMethod) Equal(other *GoMethod) bool {
	if this.Name != other.Name {
		return false
	}

	if len(this.Args) != len(other.Args) {
		return false
	}

	if len(this.Rets) != len(other.Rets) {
		return false
	}

	for i, argType := range this.Args {
		if argType.Type != other.Args[i].Type {
			return false
		}
	}

	for i, retType := range this.Rets {
		if retType.Type != other.Rets[i].Type {
			return false
		}
	}

	return true
}

func (this *GoStruct) String() string {
	var ret string = "[STRUCT] " + this.Name
	var with bool = false
	if len(this.Methods) != 0 {
		ret += " WITH METHODS: "
		with = true
		for key, _ := range this.Methods {
			ret += key + " "
		}
	}
	if len(this.Extends) != 0 {
		if with {
			ret += " AND"
		} else {
			ret += " WITH"
		}
		ret += " EXTENDS: "
		with = true
		for key, _ := range this.Extends {
			ret += key + " "
		}
	}
	if len(this.Interfaces) != 0 {
		if with {
			ret += " AND"
		} else {
			ret += " WITH"
		}
		ret += " INTERFACES: "
		for key, _ := range this.Interfaces {
			ret += key + " "
		}
	}
	return ret
}

func (this *GoInterface) String() string {
	var ret string = "[INTERFACE] " + this.Name + " WITH METHODS: "
	for key, _ := range this.Methods {
		ret += key + " "
	}
	return ret
}

func (this *GoAlias) String() string {
	var ret string = "[ALIAS] " + this.Name + " OF " + this.Type
	var with bool = false
	if len(this.Methods) != 0 {
		ret += " WITH METHODS: "
		with = true
		for key, _ := range this.Methods {
			ret += key + " "
		}
	}
	if len(this.Interfaces) != 0 {
		if with {
			ret += " AND"
		} else {
			ret += " WITH"
		}
		ret += " INTERFACES: "
		for key, _ := range this.Interfaces {
			ret += key + " "
		}
	}
	return ret
}

func (this *GoBuiltin) String() string {
	return this.Name
}
