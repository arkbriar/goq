// Copyright 2016 ArkBriar. All rights reserved.
package codelib

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/jmcvetta/neoism"
	/*
	 *"time"
	 */)

var Gopath = strings.Split(os.Getenv("GOPATH"), ":")

var _GoDB *neoism.Database

var _ObjNodeMap map[*ast.Object]*neoism.Node

var _ObjectMethodsMap map[*ast.Object][]*ast.FuncDecl

func fail(s string, a ...interface{}) {

}

func init() {
	_GoDB = nil
	_ObjNodeMap = make(map[*ast.Object]*neoism.Node)

	_ObjectMethodsMap = make(map[*ast.Object][]*ast.FuncDecl)
}

func SetDB(db *neoism.Database) {
	_GoDB = db
}

func _ExportPackageInFileToDB(file *ast.File, fnode *neoism.Node) (err error) {
	package_node, err := _GoDB.CreateNode(neoism.Props{"name": file.Name.Name})
	if err != nil {
		return err
	}

	err = package_node.AddLabel("PACKAGE")
	if err != nil {
		return err
	}

	_, err = fnode.Relate("IN_PACKAGE", package_node.Id(), neoism.Props{})

	return nil
}

func _ExportImportsInFileToDB(file *ast.File, fnode *neoism.Node) (err error) {
	// []*ast.ImportSpec
	/*
				type ImportSpec struct {
					Doc     *CommentGroup // associated documentation; or nil
		        	Name    *Ident        // local package name (including "."); or nil
		        	Path    *BasicLit     // import path
		        	Comment *CommentGroup // line comments; or nil
		        	EndPos  token.Pos     // end of spec (overrides Path.Pos if nonzero)
				}
	*/
	imports := file.Imports

	for _, import_spec := range imports {
		path := import_spec.Path
		// Assert path.Kind == token.STRING
		import_spec_node, err := _GoDB.CreateNode(neoism.Props{"name": path.Value})
		if err != nil {
			return err
		}
		err = import_spec_node.AddLabel("IMPORT")
		if err != nil {
			return err
		}

		_, err = fnode.Relate("IMPORTS", import_spec_node.Id(), neoism.Props{})
		if err != nil {
			return err
		}
	}

	return nil
}

func _ExportMethodsOfType(file *ast.File, type_spec_object *ast.Object, tnode *neoism.Node) (err error) {
	if _, ok := _ObjectMethodsMap[type_spec_object]; !ok {
		return nil
	}

	for _, method := range _ObjectMethodsMap[type_spec_object] {

		var method_node *neoism.Node = nil
		if method_node, err = _GoDB.CreateNode(neoism.Props{"name": method.Name.Name}); err != nil{
			return err
		}

		// function
		if err := method_node.AddLabel("FUNCTION"); err != nil {
			return err
		}

		if _, err := tnode.Relate("HAS", method_node.Id(), neoism.Props{}); err != nil {
			return err
		}
	}

	return nil
}


// _ExportTypesInFileToDB exports both types and interfaces
func _ExportTypesInFileToDB(file *ast.File, fnode *neoism.Node) (err error) {
	/**
	type File struct {
		Doc        *CommentGroup   // associated documentation; or nil
		Package    token.Pos       // position of "package" keyword
		Name       *Ident          // package name
		Decls      []Decl          // top-level declarations; or nil
		Scope      *Scope          // package scope (this file only)
		Imports    []*ImportSpec   // imports in this file
		Unresolved []*Ident        // unresolved identifiers in this file
		Comments   []*CommentGroup // list of all comments in the source file
	}

	type Scope struct {
		Outer   *Scope
		Objects map[string]*Object
	}
	*/

	// Scope.Outer equals to nil when it's in File

	for object_name, object := range file.Scope.Objects {
		/* Kind = Bad, Pkg, Con, Typ, Var, Fun, Lbl */
		if object.Kind != ast.Typ { // skip
			continue
		}

		var object_decl *ast.TypeSpec
		if object_decl, ok := object.Decl.(*ast.TypeSpec); ok {
			// do nothing
			object_decl.Pos();
		} else {
			panic("should not reach here.")
		}

		type_spec_node, err := _GoDB.CreateNode(neoism.Props{"name": object_name})

		if err != nil {
			return err
		}

		switch object_decl.Type.(type) {
		case *ast.ArrayType:
			err = type_spec_node.AddLabel("ARRAY_TYPE")
		case *ast.ChanType:
			err = type_spec_node.AddLabel("CHAN_TYPE")
		case *ast.FuncType:
			err = type_spec_node.AddLabel("FUNC_TYPE")
		case *ast.InterfaceType:
			err = type_spec_node.AddLabel("INTERFACE_TYPE")
		case *ast.MapType:
			err = type_spec_node.AddLabel("MAP_TYPE")
		case *ast.StructType:
			err = type_spec_node.AddLabel("STRUCT_TYPE")
		default:
			panic("oh god, how do i get here.")
		}

		if err != nil {
			return err
		}

		if err = type_spec_node.AddLabel("TYPE"); err != nil {
			return err
		}

		if _, err = fnode.Relate("DEFINES", type_spec_node.Id(), neoism.Props{}); err != nil {
			return err
		}

		if err = _ExportMethodsOfType(file, object, type_spec_node); err != nil {
			return err
		}
	}

	return nil
}

func _ExportFunctionsInFileToDB(file *ast.File, fnode *neoism.Node) (err error) {
	/* Decls:
	BadDecl,
	FuncDecl,
	GenDecl ( represents an import, constant, type or variable declaration )
	*/
	for _, decl := range file.Decls {
		var func_decl *ast.FuncDecl = nil
		switch decl.(type) {
		case *ast.FuncDecl:
			func_decl = decl.(*ast.FuncDecl)
		case *ast.BadDecl:
			continue
		case *ast.GenDecl:
			continue
		default:
			panic("should not reach here")
		}

		// assert func_decl != nil

		if func_decl.Recv == nil {
			var func_decl_node *neoism.Node

			if func_decl_node, err = _GoDB.CreateNode(neoism.Props{"name": func_decl.Name.Name}); err != nil{
				return err
			}

			// function
			if err := func_decl_node.AddLabel("FUNCTION"); err != nil {
				return err
			}

			if _, err := fnode.Relate("HAS", func_decl_node.Id(), neoism.Props{}); err != nil {
				return err
			}
		} else {
			// method
			if func_decl.Recv.NumFields() == 1 {
				field := func_decl.Recv.List[0]

				var recv_object *ast.Object = nil

				switch field.Type.(type) {
				case *ast.StarExpr:
					recv_object = field.Type.(*ast.StarExpr).X.(*ast.Ident).Obj
				case *ast.Ident:
					recv_object = field.Type.(*ast.Ident).Obj
				default:
					panic("should not reach here")
				}

				if recv_object != nil {
					if _, ok := _ObjectMethodsMap[recv_object]; !ok {
						_ObjectMethodsMap[recv_object] = make([]*ast.FuncDecl, 0, 8)
					}

					_ObjectMethodsMap[recv_object] = append(_ObjectMethodsMap[recv_object], func_decl)
				}
			}
		}
	}

	return nil
}

func _ExportFileToDB(file *ast.File, fnode *neoism.Node) (err error) {

	if err = _ExportPackageInFileToDB(file, fnode); err != nil {
		return err
	}

	if err = _ExportImportsInFileToDB(file, fnode); err != nil {
		return err
	}

	if err = _ExportFunctionsInFileToDB(file, fnode); err != nil {
		return err
	}

	if err = _ExportTypesInFileToDB(file, fnode); err != nil {
		return err
	}

	return nil
}

func _ExportDirToDB(pkgs *map[string]*ast.Package, fnode *neoism.Node) (err error) {

	return nil
}

// ExportImportsInFileToDB export imports in golang source file to neo4j.
// It will return any error it meets.
func ExportImportsInFileToDB(filename string) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	// The file set will record the position information of file
	fset := token.NewFileSet()

	// Parse the file
	f, err := parser.ParseFile(fset, filename, file, parser.ImportsOnly)
	if err != nil {
		return err
	}

	// Export FILE node
	fnode, err := _GoDB.CreateNode(neoism.Props{"name": filename})
	if err != nil {
		return err
	}
	if err = fnode.AddLabel("FILE"); err != nil {
		return err
	}

	if err = _ExportPackageInFileToDB(f, fnode); err != nil {
		return err
	}

	if err = _ExportImportsInFileToDB(f, fnode); err != nil {
		return err
	}

	return nil
}

// ExportFileToDB export all ast nodes in golang source file to neo4j.
// It will return any error it meets.
func ExportFileToDB(filename string) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	// The file set will record the position information of file
	fset := token.NewFileSet()

	// Parse the file
	f, err := parser.ParseFile(fset, filename, file, 0)
	if err != nil {
		return err
	}

	// Export FILE node
	fnode, err := _GoDB.CreateNode(neoism.Props{"name": filename})
	if err != nil {
		return err
	}
	if err = fnode.AddLabel("FILE"); err != nil {
		return err
	}

	if err = _ExportFileToDB(f, fnode); err != nil {
		return err
	}

	return nil
}

// ExportFileToDB export all ast nodes in golang source files in the given path to neo4j.
// It will return any error it meets.
func ExportDirToDB(path string) (err error) {
	// The file set will record the position information of files
	fset := token.NewFileSet()

	// Parse the dir
	_, err = parser.ParseDir(fset, path, nil, 0)
	if err != nil {
		return err
	}

	return nil
}
