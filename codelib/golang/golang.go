// Copyright 2016 ArkBriar. All rights reserved.
package golang

import (
	"go/ast"
	"go/parser"
	"go/token"

	"os"
)

type GoProject struct {
	Packages map[string]*GoPackage
}

func CreateGoProject() *GoProject {
	return &GoProject{Packages: make(map[string]*GoPackage)}
}

func __ResolveType(typeSpec *ast.TypeSpec, gfile *GoFile) {
	//@TODO
}

func __GenerateGoFileFromAstFile(astFile *ast.File, name string) (*GoFile, error) {
	var gfile *GoFile = CreateGoFile(name)

	// package name
	gfile.Package = astFile.Name.Name

	// imports)
	for _, __import := range astFile.Imports {
		gfile.Imports = append(gfile.Imports, __import.Name.Name)
	}

	// language entities: package, constant, type, variable, function(model), label
	// in ast.File.Scope.Objects, but there are no models

	for _, obj := range astFile.Scope.Objects {
		if obj.Kind != ast.Typ {
			continue
		}

		if typeSpec, ok := obj.Decl.(*ast.TypeSpec); ok {
			__ResolveType(typeSpec, gfile)
		} else {
			panic("golang/golang.go ## __GenerateGoFileFromAstFile: should not reach here")
		}
	}

	// methods

	return gfile, nil
}

func ParseFile(file *os.File) (*GoFile, error) {
	// the file set will record the postion information of file
	fset := token.NewFileSet()

	// parse the file
	var err error
	var astFile *ast.File
	if astFile, err = parser.ParseFile(fset, file.Name(), file, 0); err != nil {
		return nil, err
	}

	var gfile *GoFile
	if gfile, err = __GenerateGoFileFromAstFile(astFile, file.Name()); err != nil {
		return nil, err
	}

	return gfile, nil
}

func __ParsePackage(path string) (*GoPackage, error) {
	//@TODO

	return nil, nil
}

func ParseProject(path string) (map[string]*GoPackage, error) {
	//@TODO

	return nil, nil
}
