// Copyright 2016 ArkBriar. All rights reserved.
package golang

import (
	"go/ast"
	"go/parser"
	"go/token"

	"fmt"
	"os"
	"path"
)

func __assert(condition bool) {
	if !condition {
		panic("Assertion failed!")
	}
}

type GoProject struct {
	Name string
	Packages map[string]*GoPackage
}

func CreateGoProject(name string) *GoProject {
	return &GoProject{Name: name, Packages: make(map[string]*GoPackage)}
}

func __GetFieldTypeName(x ast.Expr) string {
	var ret string
	switch x.(type) {
	case *ast.StarExpr:
		ret = "*" + x.(*ast.StarExpr).X.(*ast.Ident).Name
	case *ast.Ident:
		ret = x.(*ast.Ident).Name
	default:
		panic("golang/golang.go ## __GetFieldTypeName: should not reach here")
	}
	return ret
}

func __ResolveStructType(structType *ast.StructType, __struct *GoStruct) {
	for _, field := range structType.Fields.List {
		typeName := __GetFieldTypeName(field.Type)
		if field.Names == nil {
			// anonymous
			__struct.__Anonymous = append(__struct.__Anonymous, typeName)
		} else {
			varName := field.Names[0].Name
			__struct.Vars[varName] = &GoVar{Name: varName, Type: typeName}
		}
	}
}

func __ResolveInterfaceType(interfaceType *ast.InterfaceType, __interface *GoInterface) {
	for _, field := range interfaceType.Methods.List {
		__assert(field.Names != nil)
		funcName := field.Names[0].Name
		var funcType *ast.FuncType = nil
		var ok bool
		if funcType, ok = field.Type.(*ast.FuncType); !ok {
			panic("golang/golang.go ## __ResolveInterfaceType: should not reach here")
		}

		__function := CreateGoFunc(funcName)
		__ResolveFuncType(funcType, __function)

		// add this method
		__interface.AddMethod(&GoMethod{GoFunc: *__function})
	}
}

func __ResolveFuncType(funcType *ast.FuncType, __function *GoFunc) {
	for _, paramField := range funcType.Params.List {
		__assert(paramField.Names != nil)
		arg := &GoVar{Name: paramField.Names[0].Name, Type: __GetFieldTypeName(paramField.Type)}
		__function.Args = append(__function.Args, arg)
	}

	for _, resultField := range funcType.Results.List {
		var ret *GoVar
		if resultField.Names != nil {
			ret = &GoVar{Name: resultField.Names[0].Name, Type: __GetFieldTypeName(resultField.Type)}
		} else {
			ret = &GoVar{Type: __GetFieldTypeName(resultField.Type)}
		}
		__function.Rets = append(__function.Rets, ret)
	}
}

func __ResolveType(typeSpec *ast.TypeSpec, gfile *GoFile) {
	Name := typeSpec.Name.Name

	switch typeSpec.Type.(type) {
	case *ast.Ident:
		fmt.Fprintln(os.Stderr, Name+" is skipped.")
	case *ast.ParenExpr:
		fmt.Fprintln(os.Stderr, Name+" is skipped.")
	case *ast.SelectorExpr:
		fmt.Fprintln(os.Stderr, Name+" is skipped.")
	case *ast.StarExpr:
		fmt.Fprintln(os.Stderr, Name+" is skipped.")
	case *ast.ArrayType:
		fmt.Fprintln(os.Stderr, Name+" is skipped.")
	case *ast.ChanType:
		fmt.Fprintln(os.Stderr, Name+" is skipped.")
	case *ast.MapType:
		/* lay aside these branches */
		//@TODO
		fmt.Fprintln(os.Stderr, Name+" is skipped.")

	case *ast.StructType:
		__struct := CreateGoStruct(Name)
		__ResolveStructType(typeSpec.Type.(*ast.StructType), __struct)
		_ = gfile.Ns.AddType(CreateGoTypeOfStruct(__struct))

	case *ast.InterfaceType:
		__interface := CreateGoInterface(Name)
		__ResolveInterfaceType(typeSpec.Type.(*ast.InterfaceType), __interface)
		_ = gfile.Ns.AddType(CreateGoTypeOfInterface(__interface))

	case *ast.FuncType:
		__function := CreateGoFunc(Name)
		__ResolveFuncType(typeSpec.Type.(*ast.FuncType), __function)
		_ = gfile.Ns.AddFunc(__function)

	}
}

func __ResolveAllMethods(astFile *ast.File, gfile *GoFile) {
	/*
		Decls:
		BadDecl,
		FuncDecl,
		GenDecl ( represents an import, constant, type or variable declaration )
	*/

	for _, decl := range astFile.Decls {
		var funcDecl *ast.FuncDecl = nil
		switch decl.(type) {
		case *ast.FuncDecl:
			funcDecl = decl.(*ast.FuncDecl)
		case *ast.GenDecl:
			continue
		case *ast.BadDecl:
			continue
		}

		__assert(funcDecl != nil)

		// functions filter
		if funcDecl.Recv == nil {
			continue
		}

		if funcDecl.Recv.NumFields() == 1 {
			field := funcDecl.Recv.List[0]
			var recvTypeName string

			// assume there are only ast.StarExpr & ast.Ident
			switch field.Type.(type) {
			case *ast.StarExpr:
				recvTypeName = field.Type.(*ast.StarExpr).X.(*ast.Ident).Name
			case *ast.Ident:
				recvTypeName = field.Type.(*ast.Ident).Name
			default:
				panic("golang/golang.go ## __GenerateGoFileFromAstFile: should not reach here")
			}

			methodName := funcDecl.Name.Name
			funcType := funcDecl.Type

			thisMethod := CreateGoMethod(methodName)

			// add functions' returns & params

			__ResolveFuncType(funcType, &thisMethod.GoFunc)

			__type := gfile.Ns.GetType(recvTypeName)

			__assert(__type.Kind == Stt || __type.Kind == Als)

			if __type.Kind == Stt { // struct
				__StructType := __type.Type.(*GoStruct)
				__StructType.AddMethod(thisMethod)
			} else { // alias
				__AliasType := __type.Type.(*GoAlias)
				__AliasType.AddMethod(thisMethod)
			}
		}
	}
}

func __IsInterfaceImplemented(methods map[string]*GoMethod, __interface *GoInterface) bool {
	for name, imethod := range __interface.Methods {
		if method, ok := methods[name]; !ok { // not found
			break
		} else {
			if method.Equal(imethod) {
				return true
			}
		}
	}

	return false
}

func __ResolveAllRelations(gfile *GoFile) {
	// pick out the structs & interfaces from anonymous that this file knowns
	for _, __struct := range gfile.Ns.GetStructs() {
		for _, anonymous := range __struct.__Anonymous {
			__a_type := gfile.Ns.GetType(anonymous)
			switch __a_type.Kind {
			case Stt:
				__struct.Extends[anonymous] = __a_type.Type.(*GoStruct)
			case Itf:
				__struct.Interfaces[anonymous] = __a_type.Type.(*GoInterface)
			case Als:
			case Bti:
			default:
				panic("golang/golang.go ## __ResolveAllRelations: should not reach here")
			}
		}
	}

	// find out interfaces implemented by type

	for _, __type := range gfile.Ns.GetTypes() {
		if __type.Kind != Stt || __type.Kind != Als {
			continue
		}

		var Methods map[string]*GoMethod = nil
		var Interfaces map[string]*GoInterface = nil

		if __type.Kind == Stt {
			Methods = __type.Type.(*GoStruct).Methods
			Interfaces = __type.Type.(*GoStruct).Interfaces
		} else {
			Methods = __type.Type.(*GoAlias).Methods
			Interfaces = __type.Type.(*GoAlias).Interfaces
		}

		__assert(Methods != nil && Interfaces != nil)

		for _, __interface := range gfile.Ns.GetInterfaces() {
			if _, ok := Interfaces[__interface.Name]; ok {
				// skip
				continue
			}
			if __IsInterfaceImplemented(Methods, __interface) {
				Interfaces[__interface.Name] = __interface
			}
		}

	}
}

func __GenerateGoFileFromAstFile(astFile *ast.File, name string) *GoFile {
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

	// methods of struct and alias
	__ResolveAllMethods(astFile, gfile)

	__ResolveAllRelations(gfile)

	return gfile
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

	gfile := __GenerateGoFileFromAstFile(astFile, file.Name())

	return gfile, nil
}

func __MergePackageFiles(gpkg *GoPackage) {
	//@TODO

}

func __ParsePackage(pkg *ast.Package, relativePath string) *GoPackage {
	gpkg := CreateGoPackage(pkg.Name, relativePath)

	for fileName, file := range pkg.Files {
		gpkg.Files[fileName] = __GenerateGoFileFromAstFile(file, fileName)
	}

	__MergePackageFiles(gpkg)

	return gpkg
}

func ParseProject(__dir string) (*GoProject, error) {
	proName := path.Base(__dir)

	fset := token.NewFileSet()

	var err error = nil
	var pkgs map[string]*ast.Package = nil

	if pkgs, err = parser.ParseDir(fset, __dir, nil, 0); err != nil {
		return nil, err
	}

	gpro := CreateGoProject(proName)

	gpkgs := gpro.Packages

	for packageName, pkg := range pkgs {
		// @TODO relative path is not considered
		gpkgs[packageName] = __ParsePackage(pkg, "")
	}

	return gpro, nil
}
