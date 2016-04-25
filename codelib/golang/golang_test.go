// Copyright 2016 ArkBriar. All rights reserved.
package golang

import (
	"testing"
	"os"
)

const (
	testdir = "testcases"
	filecase1 = "types.go"
	filecase2 = "ast.go"
	pkgcase1 = "pkgtest"
)

func __TestParseFile(t *testing.T, file string) *GoFile{
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
}

func TestParseFile2(t *testing.T) {
	__TestParseFile(t, "ast.go")
}

func TestParRePackage(t *testing.T) {
	if gpro, err := ParseProject(testdir + "/" + pkgcase1); err != nil {
		t.Fatal(err)
	} else {
		for _, pkg := range gpro.Packages {
			t.Log(pkg)
		}
	}
}
