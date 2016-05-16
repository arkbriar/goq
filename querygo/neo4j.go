// Copyright 2016 ArkBriar. All rights reserved.
// Copyright 2016 ArkBriar. All rights reserved.
package querygo

import (
	"querygo/golang"

	"github.com/jmcvetta/neoism"
)

var DB *neoism.Database

func ConnectNeo4j(username, password, url string) (db *neoism.Database, err error) {
	return neoism.Connect("http://" + username + ":" + password + "@" + url)
}

func __Connect(username, password, url string) (db *neoism.Database, err error) {
	return neoism.Connect("http://" + username + ":" + password + "@" + url)
}

type Neo4j interface {
	RollBack(db *neoism.Database) error
}

type Neo4jMap interface {
	Write(db *neoism.Database) (root *neoism.Node, first error)
}

type Neo4jNode interface {
	CreateNode(db *neoism.Database) (node *neoism.Node, first error)
}

type gopkg golang.GoPackage
type gopro golang.GoProject

func (this *gopro) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{"name": this.Name}); first != nil {
		return nil, first
	} else {
		first = node.AddLabel("PROJECT")
	}
	return node, first
}

func (this *gopro) RollBack(db *neoism.Database) error {
	return nil
}

func (this *gopro) Write(db *neoism.Database) (root *neoism.Node, first error) {
	if root, first = this.CreateNode(db); first != nil {
		return nil, first
	}

	__assert(root != nil, "querygo/neo4j.go ## gopro.Write: root should not be nil, something wrong with creation")

	// writing this project
	for _, _pkg := range this.Packages {
		pkg := gopkg(*_pkg)
		if pkgNode, err := pkg.Write(db); err != nil {
			// if error occurs, then rollback
			if _err := this.RollBack(db); _err != nil {
				return nil, _err
			}
			return nil, err
		} else {
			// Project -CONTAIN-> Package
			if _, err := root.Relate("CONTAIN", pkgNode.Id(), neoism.Props{}); err != nil {
				// if error occurs, then rollback
				if _err := this.RollBack(db); _err != nil {
					return nil, _err
				}
				return nil, err
			}
		}
	}

	// writing sub projects of this
	for _, _subpro := range this.SubPros {
		subpro := ConvertGoXxxIntoNeo4jMap(_subpro)
		if subproNode, err := subpro.Write(db); err != nil {
			return root, err
		} else {
			// Porject -HAS-> Project
			if _, err := root.Relate("HAS", subproNode.Id(), neoism.Props{}); err != nil {
				return root, err
			}
		}
	}

	return root, nil
}

func (this *gopkg) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{"name": this.Name, "path": this.RelativePath}); first != nil {
		return nil, first
	} else {
		first = node.AddLabel("PACKAGE")
	}
	return node, first
}

func (this *gopkg) Write(db *neoism.Database) (root *neoism.Node, first error) {
	if root, first = this.CreateNode(db); first != nil {
		return nil, first
	}

	__assert(root != nil, "querygo/neo4j.go ## gopkg.Write: root should not be nil, something wrong with creation")

	var NODES map[interface{}]*neoism.Node = make(map[interface{}]*neoism.Node)

	// process files one by one
	for _, _file := range this.Files {
		file := gofile(*_file)
		if fileNode, err := file.WriteWithoutRelations(db, NODES); err != nil {
			return root, err
		} else {
			if _, err := root.Relate("HAS", fileNode.Id(), neoism.Props{}); err != nil {
				return root, err
			}
			if _, err := fileNode.Relate("IN", root.Id(), neoism.Props{}); err != nil {
				return root, err
			}
		}
	}

	// process relations
	for _, _file := range this.Files {
		file := gofile(*_file)
		if err := file.WriteRelations(db, NODES); err != nil {
			return root, err
		}
	}

	return root, nil
}

func (this *gofile) WriteImports(db *neoism.Database, root *neoism.Node) (first error) {
	// imports
	for _, __import := range this.Imports {
		_import := goimport(__import)
		if importNode, err := _import.CreateNode(db); err != nil {
			return err
		} else {
			if _, err := root.Relate("IMPORT", importNode.Id(), neoism.Props{}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (this *gofile) WriteFunctions(db *neoism.Database, root *neoism.Node) (first error) {
	// functions
	for _, _function := range this.Ns.GetFuncs() {
		function := gofunc(*_function)
		if funcNode, err := function.CreateNode(db); err != nil {
			return err
		} else {
			if _, err := root.Relate("DEFINE", funcNode.Id(), neoism.Props{}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (this *gofile) WriteTypes(db *neoism.Database, root *neoism.Node, NODES map[interface{}]*neoism.Node) (first error) {
	__ProcessMethods := func(methods map[string]*golang.GoMethod, node *neoism.Node) error {
		for _, _method := range methods {
			method := gomethod(*_method)
			if methodNode, err := method.CreateNode(db); err != nil {
				return err
			} else {
				if _, err := node.Relate("HAS", methodNode.Id(), neoism.Props{}); err != nil {
					return err
				}
			}
		}

		return nil
	}

	// interfaces
	for _, __interface := range this.Ns.GetInterfaces() {
		_interface := gointerface(*__interface)

		// if not found in NODES, then create it
		var interfaceNode *neoism.Node = nil
		var err error = nil
		var ok bool = false

		if interfaceNode, ok = NODES[__interface]; !ok {
			if interfaceNode, err = _interface.CreateNode(db); err != nil {
				return err
			} else {
				// store interfaces of this package
				NODES[__interface] = interfaceNode
			}
		}

		if _, err := root.Relate("DEFINE", interfaceNode.Id(), neoism.Props{}); err != nil {
			return err
		}

		// methods
		if err := __ProcessMethods(_interface.Methods, interfaceNode); err != nil {
			return err
		}
	}

	// types
	for _, _type := range this.Ns.GetTypes() {
		if _, ok := NODES[_type.Type]; ok { // this type is already created
			continue
		}
		switch _type.Kind {
		case golang.Stt:
			__struct := _type.Type.(*golang.GoStruct)
			_struct := gostruct(*__struct)
			if structNode, err := _struct.CreateNode(db); err != nil {
				return err
			} else {
				NODES[_type.Type] = structNode
				if _, err := root.Relate("DEFINE", structNode.Id(), neoism.Props{}); err != nil {
					return err
				}
				// then methods
				if err := __ProcessMethods(_struct.Methods, structNode); err != nil {
					return err
				}
			}

		case golang.Als:
			__alias := _type.Type.(*golang.GoAlias)
			_alias := goalias(*__alias)
			if aliasNode, err := _alias.CreateNode(db); err != nil {
				return err
			} else {
				NODES[_type.Type] = aliasNode
				if _, err := root.Relate("DEFINE", aliasNode.Id(), neoism.Props{}); err != nil {
					return err
				}
				// methods
				if err := __ProcessMethods(_alias.Methods, aliasNode); err != nil {
					return err
				}
			}

		//case golang.Itf:
		//case golang.Bti:
		default:
			__assert(true, "")
		}
	}

	return nil
}

// NODES stores all types (structs | alias | interface) corresponding nodes in neo4j
func (this *gofile) WriteWithoutRelations(db *neoism.Database, NODES map[interface{}]*neoism.Node) (root *neoism.Node, first error) {
	if root, first = this.CreateNode(db); first != nil {
		return nil, first
	}

	__assert(root != nil, "querygo/neo4j.go ## gofile.WriteWithoutRelations: root should not be nil, something wrong with creation")

	if err := this.WriteImports(db, root); err != nil {
		return root, err
	}

	if err := this.WriteFunctions(db, root); err != nil {
		return root, err
	}

	if err := this.WriteTypes(db, root, NODES); err != nil {
		return root, err
	}

	return root, nil
}

func (this *gofile) WriteRelations(db *neoism.Database, NODES map[interface{}]*neoism.Node) (first error) {
	// processing interface extends
	for _, __interface := range this.Ns.GetInterfaces() {

		// when there's only one extend and no other methods, the two interface are equal
		var RELATIONSHIP string
		if len(__interface.Extends) == 1 && len(__interface.Methods) == 0 {
			RELATIONSHIP = "EQUAL_TO"
		} else {
			RELATIONSHIP = "EXTEND"
		}
		for _, _extend := range __interface.Extends {
			if interfaceNode, ok := NODES[__interface]; !ok {
				panic("querygo/neo4j.go ## gfile.Write: should not reach here")
			} else {
				if extendNode, ok := NODES[_extend]; !ok {
					panic("querygo/neo4j.go ## gfile.Write: should not reach here")
				} else {
					if _, err := interfaceNode.Relate(RELATIONSHIP, extendNode.Id(), neoism.Props{}); err != nil {
						return err
					}
				}
			}
		}
	}

	__ProcessImplements := func(interfaces map[string]*golang.GoInterface, node *neoism.Node) error {
		for _, __interface := range interfaces {
			_interface := gointerface(*__interface)
			var interfaceNode *neoism.Node = nil
			var err error = nil
			var ok bool = false
			// if there's no before, we should create it.
			if interfaceNode, ok = NODES[__interface]; !ok {
				if interfaceNode, err = _interface.CreateNode(db); err != nil {
					return err
				} else {
					NODES[__interface] = interfaceNode
				}
			}
			if _, err := node.Relate("IMPLEMENT", interfaceNode.Id(), neoism.Props{}); err != nil {
				return err
			}
		}

		return nil
	}

	// processing struct's extends & implements
	for _, __struct := range this.Ns.GetStructs() {
		if structNode, ok := NODES[__struct]; !ok {
			panic("querygo/neo4j.go ## gfile.Write: should not reach here")
		} else {
			for _, _extend := range __struct.Extends {
				if extendNode, ok := NODES[_extend]; !ok {
					panic("querygo/neo4j.go ## gfile.Write: should not reach here")
				} else {
					if _, err := structNode.Relate("EXTEND", extendNode.Id(), neoism.Props{}); err != nil {
						return err
					}
				}
			}
			if err := __ProcessImplements(__struct.Interfaces, structNode); err != nil {
				return err
			}
		}
	}

	// processing alias' implements
	for _, __alias := range this.Ns.GetAliases() {
		if aliasNode, ok := NODES[__alias]; !ok {
			panic("querygo/neo4j.go ## gfile.Write: should not reach here")
		} else {
			if err := __ProcessImplements(__alias.Interfaces, aliasNode); err != nil {
				return err
			}
		}
	}

	return nil
}

func (this *gofile) Write(db *neoism.Database) (root *neoism.Node, first error) {
	// store types (structs) and interfaces
	var NODES map[interface{}]*neoism.Node = make(map[interface{}]*neoism.Node)

	if root, first = this.WriteWithoutRelations(db, NODES); first != nil {
		return
	}

	if first = this.WriteRelations(db, NODES); first != nil {
		return
	}

	return root, nil
}

type (
	gofile      golang.GoFile
	goalias     golang.GoAlias
	gostruct    golang.GoStruct
	goimport    string
	gointerface golang.GoInterface
	gofunc      golang.GoFunc
	gomethod    golang.GoMethod
)

// should always called with *GoFile, *GoPackage and *GoProject, otherwise it will trigger a panic
func ConvertGoXxxIntoNeo4jMap(goxxx interface{}) Neo4jMap {
	var ret Neo4jMap = nil
	switch goxxx.(type) {
	case *golang.GoFile:
		tmp := gofile(*goxxx.(*golang.GoFile))
		ret = &tmp
	case *golang.GoPackage:
		tmp := gopkg(*goxxx.(*golang.GoPackage))
		ret = &tmp
	case *golang.GoProject:
		tmp := gopro(*goxxx.(*golang.GoProject))
		ret = &tmp
	default:
		__assert(false, "querygo/neo4j.go ## ConvertGoXxxIntoNeo4jMap: should not reach here, check your code.")
	}

	return ret
}

// should always called with *GoXxx or string, otherwise it will trigger a panic
func ConvertGoXxxIntoNeo4jNode(goxxx interface{}) Neo4jNode {
	var ret Neo4jNode = nil
	switch goxxx.(type) {
	case *golang.GoFile:
		tmp := gofile(*goxxx.(*golang.GoFile))
		ret = &tmp
	case *golang.GoPackage:
		tmp := gopkg(*goxxx.(*golang.GoPackage))
		ret = &tmp
	case *golang.GoProject:
		tmp := gopro(*goxxx.(*golang.GoProject))
		ret = &tmp
	case *golang.GoStruct:
		tmp := gostruct(*goxxx.(*golang.GoStruct))
		ret = &tmp
	case *golang.GoAlias:
		tmp := goalias(*goxxx.(*golang.GoAlias))
		ret = &tmp
	case string:
		tmp := goimport(goxxx.(string))
		ret = &tmp
	case *golang.GoInterface:
		tmp := gointerface(*goxxx.(*golang.GoInterface))
		ret = &tmp
	case *golang.GoFunc:
		tmp := gofunc(*goxxx.(*golang.GoFunc))
		ret = &tmp
	case *golang.GoMethod:
		tmp := gomethod(*goxxx.(*golang.GoMethod))
		ret = &tmp
	default:
		__assert(false, "querygo/neo4j.go ## ConvertGoXxxIntoNeo4jNode: should not reach here, check your code.")
	}

	return ret
}

func _Public(public bool) string {
	if public {
		return "public"
	}
	return "private"
}

func (this *gofile) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{"name": this.Name}); first != nil {
		return
	}

	first = node.AddLabel("FILE")
	return
}

func (this *goalias) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{
		"name":          this.Name,
		"visibility": _Public((*golang.GoAlias)(this).IsPublic()),
		"alias_of":      this.Type,
	}); first != nil {
		return
	}

	first = node.AddLabel("ALIAS", "TYPE")
	return
}

func (this *gointerface) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{
		"name":          this.Name,
		"visibility": _Public((*golang.GoInterface)(this).IsPublic()),
	}); first != nil {
		return
	}

	first = node.AddLabel("INTERFACE", "TYPE")
	return
}

func (this *gostruct) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{
		"name":          this.Name,
		"visibility": _Public((*golang.GoStruct)(this).IsPublic()),
	}); first != nil {
		return
	}

	first = node.AddLabel("STRUCT", "TYPE")
	return
}

func (this *gofunc) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{
		"name":          this.Name,
		"visibility": _Public((*golang.GoFunc)(this).IsPublic()),
	}); first != nil {
		return
	}

	first = node.AddLabel("FUNCTION")
	return
}

func (this *gomethod) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{
		"name":          this.Name,
		"visibility": _Public((*golang.GoMethod)(this).IsPublic()),
	}); first != nil {
		return
	}

	first = node.AddLabel("METHOD")
	return
}

func (this goimport) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{"name": this}); first != nil {
		return
	}

	first = node.AddLabel("IMPORT")
	return
}

func (this *gofile) RollBack(db *neoism.Database) error {

	return nil
}

func (this *goalias) RollBack(db *neoism.Database) error {

	return nil
}

func (this *gofunc) RollBack(db *neoism.Database) error {

	return nil
}

func (this *goimport) RollBack(db *neoism.Database) error {

	return nil
}

func (this *gomethod) RollBack(db *neoism.Database) error {

	return nil
}

func (this *gostruct) RollBack(db *neoism.Database) error {

	return nil
}

func (this *gointerface) RollBack(db *neoism.Database) error {

	return nil
}
