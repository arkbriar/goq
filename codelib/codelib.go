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
	 */
	"fmt"
)

var Gopath = strings.Split(os.Getenv("GOPATH"), ":")

var _GoDB *neoism.Database

var _ObjNodeMap map[*ast.Object]*neoism.Node

func fail(s string, a ...interface{}) {

}

func init() {
	_GoDB = nil
	_ObjNodeMap = make(map[*ast.Object]*neoism.Node)
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

// _ExportTypesInFileToDB exports both types and interfaces
func _ExportTypesInFileToDB(file *ast.File, fnode *neoism.Node) (err error) {

	for _, decl := range file.Decls {
		// Type Assertion
		if gendecl, ok := decl.(*ast.GenDecl); ok {
			// TOKEN: TYPE {struct | interface}
			if gendecl.Tok == token.TYPE {
				for _, spec := range gendecl.Specs {
					// Type Assertion
					switch spec_type := spec.(type) {
					case *ast.TypeSpec:
						type_spec, _ := spec.(*ast.TypeSpec)
						type_spec_node, err := _GoDB.CreateNode(neoism.Props{"name": type_spec.Name.Name})
						if err != nil {
							return err
						}
						// *Ident, *ParenExpr, *SelectorExpr, *StarExpr and *XxxTypes
						switch type_spec.Type.(type) {
						case *ast.InterfaceType:
							_ObjNodeMap[type_spec.Name.Obj] = type_spec_node
							// Struct Info
							err := type_spec_node.SetLabels([]string{"TYPE", "INTERFACE"})
							if err != nil {
								return err
							}
						case *ast.StructType:
							_ObjNodeMap[type_spec.Name.Obj] = type_spec_node
							// Interface Info
							err := type_spec_node.SetLabels([]string{"TYPE", "STRUCT"})
							if err != nil {
								return err
							}
						// Ignored
						/*
						 *case *ast.ArrayType:
						 *case *ast.ChanType:
						 *case *ast.FuncType:
						 *case *ast.MapType:
						 */
						/*
						 *case *ast.Ident:
						 *case *ast.ParenExpr:
						 *case *ast.SelectorExpr:
						 *case *ast.StarExpr:
						 */
						default:
						}
					/*
					 *case *ast.ImportSpec:
					 *case *ast.ValueSpec:
					 */
					default:
						fmt.Printf("Unexpected type %T\n", spec_type)
					}
				}
			}
		}
	}

	return nil
}

func _ExportFunctionsInFileToDB(file *ast.File, fnode *neoism.Node) (err error) {

	for _, decl := range file.Decls {
		// Type Assertion
		if _, ok := decl.(*ast.FuncDecl); ok {

		}
	}

	return nil
}

func _ExportFileToDB(file *ast.File, fnode *neoism.Node) (err error) {

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
	_, err = parser.ParseFile(fset, filename, file, 0)
	if err != nil {
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
