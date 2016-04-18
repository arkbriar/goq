// Copyright 2016 ArkBriar. All rights reserved.
package codelib

import (
	"codelib/golang"

	"github.com/jmcvetta/neoism"
)

func __Connect(username, password, url string) (db *neoism.Database, err error) {
	return neoism.Connect("http://" + username + ":" + password + "@" + url)
}

func ConnectToDB(username, password, url string) (db *neoism.Database, err error) {
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

	__assert(root != nil, "codelib/neo4j.go ## gopro.Write: root should not be nil, something wrong with creation")

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

	__assert(root != nil, "codelib/neo4j.go ## gopkg.Write: root should not be nil, something wrong with creation")

	// process files one by one
	for _, _file := range this.Files {
		file := gofile(*_file)
		if fileNode, err := file.Write(db); err != nil {
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

	return root, nil
}

func (this *gofile) Write(db *neoism.Database) (root *neoism.Node, first error) {
	if root, first = this.CreateNode(db); first != nil {
		return nil, first
	}

	__assert(root != nil, "codelib/neo4j.go ## gofile.Write: root should not be nil, something wrong with creation")

	// store types (structs and aliass)
	var NODES map[interface{}]*neoism.Node = make(map[interface{}]*neoism.Node)

	// imports
	for _, __import := range this.Imports {
		_import := goimport(__import)
		if importNode, err := _import.CreateNode(db); err != nil {
			return root, err
		} else {
			if _, err := root.Relate("IMPORT", importNode.Id(), neoism.Props{}); err != nil {
				return root, err
			}
		}
	}

	// functions
	for _, _function := range this.Ns.GetFuncs() {
		function := gofunc(*_function)
		if funcNode, err := function.CreateNode(db); err != nil {
			return root, err
		} else {
			if _, err := root.Relate("DEFINE", funcNode.Id(), neoism.Props{}); err != nil {
				return root, err
			}
		}
	}

	__ProcessMethods := func (methods map[string]*golang.GoMethod, node *neoism.Node) error {
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
		if interfaceNode, err := _interface.CreateNode(db); err != nil {
			return root, err
		} else {
			if _, err := root.Relate("DEFINE", interfaceNode.Id(), neoism.Props{}); err != nil {
				return root, err
			}

			// methods
			if err := __ProcessMethods(_interface.Methods, interfaceNode); err != nil {
				return root, err
			}
		}
	}

	__ProcessImplements := func (interfaces map[string]*golang.GoInterface, node *neoism.Node) error {
		for _, __interface := range interfaces {
			_interface := gointerface(*__interface)
			if interfaceNode, err := _interface.CreateNode(db); err != nil {
				return err
			} else {
				if _, err := node.Relate("IMPLEMENT", interfaceNode.Id(), neoism.Props{}); err != nil {
					return err
				}
			}
		}

		return nil
	}

	// types
	for _, _type := range this.Ns.GetTypes() {
		if _, ok := NODES[_type.Type]; ok { // this type is already created
			continue
		}
		switch _type.Kind {
		case golang.Stt:
			_struct := _type.Type.(*gostruct)
			if structNode, err := _struct.CreateNode(db); err != nil {
				return root, err
			} else {
				if _, err := root.Relate("DEFINE", structNode.Id(), neoism.Props{}); err != nil {
					return root, err
				}
				// then methods
				if err := __ProcessMethods(_struct.Methods, structNode); err != nil {
					return root, err
				}
				// then implements
				if err := __ProcessImplements(_struct.Interfaces, structNode); err != nil {
					return root, err
				}
				// then extends

				for _, _extend := range _struct.Extends {
					var extendNode *neoism.Node = nil
					var ok bool = false
					// if not created, then create it
					if extendNode, ok = NODES[_extend]; !ok {
						extend := gostruct(*_extend)
						var err error
						if extendNode, err = extend.CreateNode(db); err != nil {
							return root, err
						}
					}
					if _, err := structNode.Relate("EXTEND", extendNode.Id(), neoism.Props{}); err != nil {
						return root, err
					}
				}
			}

		case golang.Als:
			_alias := _type.Type.(*goalias)
			if aliasNode, err := _alias.CreateNode(db); err != nil {
				return root, err
			} else {
				if _, err := root.Relate("DEFINE", aliasNode.Id(), neoism.Props{}); err != nil {
					return root, err
				}
				// methods
				if err := __ProcessMethods(_alias.Methods, aliasNode); err != nil {
					return root, err
				}
				// implements
				if err := __ProcessImplements(_alias.Interfaces, aliasNode); err != nil {
					return root, err
				}
			}

		//case golang.Itf:
		//case golang.Bti:
		default:
			__assert(true, "")
		}
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

func (this *gofile) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{"name": this.Name}); first != nil {
		return
	}

	first = node.AddLabel("FILE")
	return
}

func (this *goalias) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{"name": this.Name}); first != nil {
		return
	}

	first = node.AddLabel("TYPE", "ALIAS")
	return
}

func (this *gointerface) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{"name": this.Name}); first != nil {
		return
	}

	first = node.AddLabel("TYPE", "INTERFACE")
	return nil, nil
}

func (this *gostruct) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{"name": this.Name}); first != nil {
		return
	}

	first = node.AddLabel("TYPE", "STRUCT")
	return
}

func (this *gofunc) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{"name": this.Name}); first != nil {
		return
	}

	first = node.AddLabel("FUNCTION")
	return
}

func (this *gomethod) CreateNode(db *neoism.Database) (node *neoism.Node, first error) {
	if node, first = db.CreateNode(neoism.Props{"name": this.Name}); first != nil {
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
