// Copyright 2016 ArkBriar. All rights reserved.
package golang

import (
	"errors"
)

type nsMethods interface {
	AddType(__type *GoType) error
	AddFunc(__func *GoFunc) error
	GetTypes() []*GoType
	GetFuncs() []*GoFunc
	GetType(name string) *GoType
	GetFunc(name string) *GoFunc
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
	if _, ok := this.Types[__type.Name()]; !ok {
		return errors.New("Type " + __type.Name() + " already exists!")
	}

	this.Types[__type.Name()] = __type

	return nil
}

func (this *goNamespace) AddFunc(__func *GoFunc) error {
	if _, ok := this.Funcs[__func.Name]; !ok {
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

func (this *goNamespace) GetFuncs() []*GoFunc {
	var funcs []*GoFunc = make([]*GoFunc, 0, len(this.Funcs))

	for _, v := range this.Funcs {
		funcs = append(funcs, v)
	}

	return funcs
}

func (this *goNamespace) GetType(name string) *GoType {
	return this.Types[name]
}

func (this *goNamespace) GetFunc(name string) *GoFunc {
	return this.Funcs[name]
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
		Name string  // var name
		Type *GoType // var type
	}

	GoStruct struct {
		Name       string                  // struct name
		Vars       map[string]*GoVar       // variables in struct
		Methods    map[string]*GoMethod    // methods
		Interfaces map[string]*GoInterface // interfaces
		Extends    map[string]*GoStruct    // extends
	}

	GoInterface struct {
		Name    string               // interface name
		Methods map[string]*GoMethod // methods
	}

	GoAlias struct {
		Name       string                  // alias name
		Type       *GoType                 // original type
		Methods    map[string]*GoMethod    // methods
		Interfaces map[string]*GoInterface // interfaces
	}

	GoBuiltin struct {
		Name string
	}

	GoFunc struct {
		Name string            // function name
		Args map[string]*GoVar // function args
		Rets []*GoVar          // function rets
	}

	GoMethod struct {
		GoFunc // extends function
		// who receives this method
		// Recv *GoStruct/GoAlias
	}
)

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

func CreateGoFile(name string) *GoFile {
	gf := &GoFile{
		Name:    name,
		Ns:      __CreateGoNameSpace(),
		Imports: make([]string, 0, 8),
	}

	return gf
}

func CreateGoStruct(name string) *GoStruct {
	gs := &GoStruct{
		Name:       name,
		Vars:       make(map[string]*GoVar),
		Methods:    make(map[string]*GoMethod),
		Interfaces: make(map[string]*GoInterface),
		Extends:    make(map[string]*GoStruct),
	}

	return gs
}

func CreateGoInterface(name string) *GoInterface {
	gi := &GoInterface{
		Name:    name,
		Methods: make(map[string]*GoMethod),
	}

	return gi
}

func CreateGoAlias(name string) *GoAlias {
	ga := &GoAlias{
		Name:       name,
		Type:       nil,
		Methods:    make(map[string]*GoMethod),
		Interfaces: make(map[string]*GoInterface),
	}

	return ga
}

func CreateGoFunc(name string) *GoFunc {
	gf := &GoFunc{
		Name: name,
		Args: make(map[string]*GoVar),
		Rets: make([]*GoVar, 0, 2),
	}

	return gf
}

func CreateGoMethod(name string) *GoMethod {
	return &GoMethod{GoFunc: *CreateGoFunc(name)}
}
