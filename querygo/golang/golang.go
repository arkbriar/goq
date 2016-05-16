// Copyright 2016 ArkBriar. All rights reserved.
package golang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"querygo/debug"

	"fmt"
	"os"
	"path"
	"path/filepath"
)

var dbg debug.DebugLog

func SetDebug(D debug.DebugLog) {
	dbg = D
}

func __assert(condition bool) {
	if !condition {
		panic("Assertion failed!")
	}
}

type GoProject struct {
	Name     string
	Packages map[string]*GoPackage
	SubPros  map[string]*GoProject
	Upper    *GoProject
}

func CreateGoProject(name string) *GoProject {
	return &GoProject{
		Name:     name,
		Packages: make(map[string]*GoPackage),
		SubPros:  make(map[string]*GoProject),
		Upper:    nil,
	}
}

func __RemoveFirstStar(x string) string {
	if x[0] == '*' {
		return x[1:]
	}
	return x
}

func __GetFieldTypeName(x ast.Expr) string {
	var ret string
	switch x.(type) {
	case *ast.StarExpr:
		ret = "*" + __GetFieldTypeName(x.(*ast.StarExpr).X)
	case *ast.Ident:
		ret = x.(*ast.Ident).Name
	case *ast.SelectorExpr:
		s := x.(*ast.SelectorExpr)
		ret = __GetFieldTypeName(s.X) + "." + s.Sel.Name
	case *ast.ArrayType:
		ret = "[]" + __GetFieldTypeName(x.(*ast.ArrayType).Elt)
	case *ast.ChanType:
		ret = "chan " + __GetFieldTypeName(x.(*ast.ChanType).Value)
	case *ast.MapType:
		m := x.(*ast.MapType)
		ret = "map[" + __GetFieldTypeName(m.Key) + "]" + __GetFieldTypeName(m.Value)
	case *ast.ParenExpr:
		ret = __GetFieldTypeName(x.(*ast.ParenExpr).X)
	case *ast.InterfaceType:
		// special case, the empty interface `interface{}`
		__assert(x.(*ast.InterfaceType).Methods.NumFields() == 0)
		ret = "interface{}"
	case *ast.FuncType:
		f := x.(*ast.FuncType)
		ret += "func("
		// vars
		var paramLen, retLen int = 0, 0
		if f.Params != nil && f.Params.List != nil {
			paramLen = len(f.Params.List)
		}
		if f.Results != nil && f.Results.List != nil {
			retLen = len(f.Results.List)
		}
		for i := 0; i < paramLen; i++ {
			ret += __GetFieldTypeName(f.Params.List[i].Type)
			if i != paramLen-1 {
				ret += ", "
			}
		}
		ret += ")"
		if retLen == 1 {
			ret += " " + __GetFieldTypeName(f.Results.List[0].Type)
		} else if retLen > 1 {
			ret += "("
			for i := 0; i < retLen; i++ {
				ret += __GetFieldTypeName(f.Results.List[i].Type)
				if i != retLen-1 {
					ret += ", "
				}
			}
			ret += ")"
		}

	case *ast.Ellipsis:
		e := x.(*ast.Ellipsis)
		ret += "[]" + __GetFieldTypeName(e.Elt)

	case *ast.StructType:
		s := x.(*ast.StructType)
		ret += "struct {\n"
		if s.Fields != nil {
			for _, field := range s.Fields.List {
				if field.Names != nil {
					ret += field.Names[0].Name + " "
				}
				ret += __GetFieldTypeName(field.Type) + "\n"
			}
		}
		ret += "}"
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
			for _, _var := range field.Names {
				varName := _var.Name
				__struct.Vars[varName] = &GoVar{Name: varName, Type: typeName}
			}
		}
	}
}

func __ResolveInterfaceType(interfaceType *ast.InterfaceType, __interface *GoInterface) {
	for _, field := range interfaceType.Methods.List {
		if field.Names == nil { // anonymous field
			__interface.__Anonymous = append(__interface.__Anonymous, __GetFieldTypeName(field.Type))
			continue
		}
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
		// could be nil, like `func(string)bool`
		var arg *GoVar = nil
		if paramField.Names != nil {
			arg = &GoVar{Name: paramField.Names[0].Name, Type: __GetFieldTypeName(paramField.Type)}
		} else {
			arg = &GoVar{Name: "", Type: __GetFieldTypeName(paramField.Type)}
		}
		__function.Args = append(__function.Args, arg)
	}

	if funcType.Results != nil {
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
}

func __ResolveType(typeSpec *ast.TypeSpec, gfile *GoFile) {
	Name := typeSpec.Name.Name

	switch typeSpec.Type.(type) {
	/*
	 *case *ast.Ident:
	 *case *ast.ParenExpr:
	 *case *ast.SelectorExpr:
	 *case *ast.StarExpr:
	 *case *ast.ArrayType:
	 *case *ast.ChanType:
	 *case *ast.MapType:
	 */

	case *ast.StructType:
		__struct := CreateGoStruct(Name)
		__ResolveStructType(typeSpec.Type.(*ast.StructType), __struct)
		_ = gfile.Ns.AddType(CreateGoTypeOfStruct(__struct))

	case *ast.InterfaceType:
		__interface := CreateGoInterface(Name)
		__ResolveInterfaceType(typeSpec.Type.(*ast.InterfaceType), __interface)
		_ = gfile.Ns.AddType(CreateGoTypeOfInterface(__interface))

	case *ast.FuncType:
		// if type == *ast.FuncType, then it must be something like `type A func(x int) bool`
		__alias := CreateGoAlias(Name, __GetFieldTypeName(typeSpec.Type))
		_ = gfile.Ns.AddType(CreateGoTypeOfAlias(__alias))

	default:
		__alias := CreateGoAlias(Name, __GetFieldTypeName(typeSpec.Type))
		_ = gfile.Ns.AddType(CreateGoTypeOfAlias(__alias))
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

			recvTypeName := __GetFieldTypeName(field.Type)

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
			return false
		} else {
			if !method.Equal(imethod) {
				return false
			}
		}
	}

	return true
}

func __ResolveAllRelations(gfile *GoFile) {
	// for interface anonymous
	for _, __interface := range gfile.Ns.GetInterfaces() {
		for _, anonymous := range __interface.__Anonymous {
			__a_type := gfile.Ns.GetType(anonymous)
			// must be interface in interface, otherwise the compiler will give an error
			__assert(__a_type.Kind == Itf)
			__a_itf := __a_type.Type.(*GoInterface)
			__interface.Extends[anonymous] = __a_itf
		}
	}

	// pick out the structs & interfaces from anonymous that this file knowns
	for _, __struct := range gfile.Ns.GetStructs() {
		for _, anonymous := range __struct.__Anonymous {
			__a_type := gfile.Ns.GetType(anonymous)
			switch __a_type.Kind {
			case Stt:
				__struct.Extends[anonymous] = __a_type.Type.(*GoStruct)
			case Itf:
				__struct.Interfaces[anonymous] = __a_type.Type.(*GoInterface)
				// delete these functions (existance ensured by compiler)
				for methodName, _ := range __a_type.Type.(*GoInterface).Methods {
					delete(__struct.Methods, methodName)
				}
			case Als:
			case Bti:
			default:
				panic("golang/golang.go ## __ResolveAllRelations: should not reach here")
			}
		}
	}

	// find out interfaces implemented by type
	for _, __type := range gfile.Ns.GetTypes() {
		if __type.Kind != Stt && __type.Kind != Als {
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
				// delete these functions (existance ensured by compiler)
				for methodName, _ := range __interface.Methods {
					delete(Methods, methodName)
				}
			}
		}
	}
	/*
	 * Type's Methods += Type's Extends' Methods + Types's Interfaces' Methods
	 * Type's Interfaces += Type's Extends's Interfaces
	 */
}

func __GenerateGoFileFromAstFile(astFile *ast.File, name string) *GoFile {
	var gfile *GoFile = CreateGoFile(name)

	// package name
	gfile.Package = astFile.Name.Name

	// imports
	for _, __import := range astFile.Imports {
		if __import.Name != nil { // local package name, include '.'
			gfile.Imports = append(gfile.Imports, __import.Name.Name)
		} else {
			gfile.Imports = append(gfile.Imports, __import.Path.Value)
		}
	}

	// language entities: package, constant, type, variable, function(model), label
	// in ast.File.Scope.Objects, but there are no models

	for _, obj := range astFile.Scope.Objects {
		if obj.Kind == ast.Typ {
			if typeSpec, ok := obj.Decl.(*ast.TypeSpec); ok {
				__ResolveType(typeSpec, gfile)
				if (obj.Data != nil) {
					fmt.Println("data isn't nil, the type is " + obj.Name)
				}
			} else {
				panic("golang/golang.go ## __GenerateGoFileFromAstFile: should not reach here")
			}
		} else if obj.Kind == ast.Fun {
			if funcDecl, ok := obj.Decl.(*ast.FuncDecl); ok {
				funcType := funcDecl.Type
				__function := CreateGoFunc(funcDecl.Name.Name)
				__ResolveFuncType(funcType, __function)
				gfile.Ns.AddFunc(__function)
			}
		}
	}

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

	// methods of struct and alias
	__ResolveAllMethods(astFile, gfile)

	__ResolveAllRelations(gfile)

	return gfile, nil
}

func (this *GoPackage) GetType(name string) *GoType {
	for _, file := range this.Files {
		if __type := file.Ns.GetType(name); __type != nil {
			return __type
		}
	}
	return nil
}

// gfile must be one file in gpkg
func __ResolveAllMethodsInPackage(astFile *ast.File, gpkg *GoPackage) {
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

			recvTypeName := __GetFieldTypeName(field.Type)

			methodName := funcDecl.Name.Name
			funcType := funcDecl.Type

			thisMethod := CreateGoMethod(methodName)

			// add functions' returns & params
			__ResolveFuncType(funcType, &thisMethod.GoFunc)

			__type := gpkg.GetType(recvTypeName)

			if __type == nil {
				//@TODO
				continue
			}

			__assert(__type != nil && (__type.Kind == Stt || __type.Kind == Als))

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

// gfile must be one file in gpkg
func __ResolveAllRelationsInPackage(gfile *GoFile, gpkg *GoPackage) {
	// for interface anonymous
	for _, __interface := range gfile.Ns.GetInterfaces() {
		for _, anonymous := range __interface.__Anonymous {
			__a_type := gpkg.GetType(anonymous)
			if __a_type == nil {
				//@TODO
				continue
			}
			// must be interface in interface, otherwise the compiler will give an error
			__assert(__a_type.Kind == Itf)
			__a_itf := __a_type.Type.(*GoInterface)
			__interface.Extends[anonymous] = __a_itf
		}
	}

	// pick out the structs & interfaces from anonymous that this file knowns
	for _, __struct := range gfile.Ns.GetStructs() {
		for _, anonymous := range __struct.__Anonymous {
			__a_type := gpkg.GetType(anonymous)
			if __a_type == nil {
				//@TODO this is a type defined in another package
				continue
			}
			switch __a_type.Kind {
			case Stt:
				__struct.Extends[anonymous] = __a_type.Type.(*GoStruct)
			case Itf:
				__struct.Interfaces[anonymous] = __a_type.Type.(*GoInterface)
				// delete these functions (existance ensured by compiler)
				for methodName, _ := range __a_type.Type.(*GoInterface).Methods {
					delete(__struct.Methods, methodName)
				}
			case Als:
			case Bti:
			default:
				panic("golang/golang.go ## __ResolveAllRelations: should not reach here")
			}

		}
	}

	// find out interfaces implemented by type
	for _, __type := range gfile.Ns.GetTypes() {
		if __type.Kind != Stt && __type.Kind != Als {
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

		for _, _gfile := range gpkg.Files {
			for _, __interface := range _gfile.Ns.GetInterfaces() {
				if _, ok := Interfaces[__interface.Name]; ok {
					// skip
					continue
				}
				if __IsInterfaceImplemented(Methods, __interface) {
					Interfaces[__interface.Name] = __interface
					// delete these functions (existance ensured by compiler)
					for methodName, _ := range __interface.Methods {
						delete(Methods, methodName)
					}
				}
			}
		}
	}
	/*
	 * Type's Methods += Type's Extends' Methods + Types's Interfaces' Methods
	 * Type's Interfaces += Type's Extends's Interfaces
	 */
}

func __MergePackageFiles(pkg *ast.Package, gpkg *GoPackage) {
	// methods of struct and alias
	for _, file := range pkg.Files {
		__ResolveAllMethodsInPackage(file, gpkg)
	}

	for _, gfile := range gpkg.Files {
		__ResolveAllRelationsInPackage(gfile, gpkg)
	}
}

func __ParsePackage(pkg *ast.Package, relativePath string) *GoPackage {
	gpkg := CreateGoPackage(pkg.Name, relativePath)

	for fileName, file := range pkg.Files {
		fileName = filepath.Join(relativePath, filepath.Base(fileName))
		/*
		 *fmt.Fprintf(os.Stdout, "Processing %s\n", fileName)
		 */
		gpkg.Files[fileName] = __GenerateGoFileFromAstFile(file, fileName)
	}

	__MergePackageFiles(pkg, gpkg)

	return gpkg
}

func __ParseDir(__dir string, __relative_path string) (*GoProject, error) {
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
		gpkgs[packageName] = __ParsePackage(pkg, __relative_path)
	}

	// parse the subdirs recursively
	const all = -1

	if df, err := os.Open(__dir); err != nil {
		return nil, err
	} else {
		if fi, err := df.Readdir(all); err == nil {
			for _, can_dir := range fi {
				if can_dir.IsDir() {
					// parse this dir
					sub_dir := __dir + "/" + can_dir.Name()
					//@TODO
					if can_dir.Name() == "testcases" {
						continue
					}
					if _gpro, err := __ParseDir(sub_dir, filepath.Join(__relative_path, can_dir.Name())); err != nil {
						return gpro, err
					} else if _gpro != nil {
						gpro.SubPros[_gpro.Name] = _gpro
						_gpro.Upper = gpro
					}
				}
			}
		}
	}

	// empty dir (there're no go src files in this dir)
	if len(gpro.Packages) == 0 && len(gpro.SubPros) == 0 {
		return nil, nil
		//return nil, errors.New("Empty dir without any go src files or dirs: " + __dir)
	} else {
		return gpro, nil
	}
}

func ParseProject(__dir string) (*GoProject, error) {
	absolute_path, err := filepath.Abs(__dir)
	if err != nil {
		return nil, err
	}
	pro, err := __ParseDir(absolute_path, "")
	return pro, err
}

// this function is only for test
func __PrintNamespace(ns *goNamespace) {
	fmt.Println(ns.Types)
	fmt.Println(ns.Funcs)
}
