// Copyright 2016 ArkBriar. All rights reserved.
package golang

import (
	"os"
	"testing"
)

const (
	testdir = "testcases"

	filecase_1 = "types.go"
	filecase_2 = "ast.go"
	pkgcase_1  = "pkgtest"
	pkgcase_2  = "ast"
)

func __TestParseFile(t *testing.T, file string) *GoFile {
	if file, err := os.Open(testdir + "/" + file); err != nil {
		t.Fatal(err)
	} else {
		defer file.Close()

		gfile, err := ParseFile(file)
		if err != nil {
			t.Fatal(err)
		}
		return gfile
	}
	return nil
}

func TestParseFile(t *testing.T) {
	gfile := __TestParseFile(t, "types.go")
	__PrintNamespace(gfile.Ns)

	__TestParseFile(t, "ast.go")
}

func TestParsePackage(t *testing.T) {
	if gpro, err := ParseProject(testdir + "/" + pkgcase_1); err != nil {
		t.Fatal(err)
	} else {
		for _, pkg := range gpro.Packages {
			t.Log(pkg)
		}
	}

	if gpro, err := ParseProject(testdir + "/" + pkgcase_2); err != nil {
		t.Fatal(err)
	} else {
		for _, pkg := range gpro.Packages {
			t.Log(pkg)
		}
	}
}

func TestSelfParse(t *testing.T) {
	if gpro, err := ParseProject("."); err != nil {
		t.Fatal(err)
	} else {
		for _, pkg := range gpro.Packages {
			t.Log(pkg)
		}
	}
}
