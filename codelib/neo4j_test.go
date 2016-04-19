// Copyright 2016 ArkBriar. All rights reserved.
package codelib

import (
	"testing"
	"codelib/golang"
	"os"
)

const (
	username = "neo4j"
	password = "dsj1994"
	url = "localhost:7474/db/data"
)

const (
	testdir = "golang/testcases"

	__case_1 = "types.go"
	__case_2 = "ast.go"
)

func __TestParseFile(t *testing.T, file string) *golang.GoFile{
	if file, err := os.Open(testdir + "/" + file); err != nil {
		t.Fatal(err)
	} else {
		defer file.Close()

		gfile, err := golang.ParseFile(file)
		if err != nil {
			t.Fatal(err)
		}
		return gfile
	}
	return nil
}

func TestExportGoPackageToNeo4j(t *testing.T) {
	if db, err := __Connect(username, password, url); err != nil {
		t.Fatal(err)
	} else {
		_gfile := __TestParseFile(t, __case_1)
		gfile := gofile(*_gfile)
		if _, err := gfile.Write(db); err != nil {
			t.Fatal(err)
		}
	}
}

func TestExportGoPackageToNeo4j2(t *testing.T) {
	if db, err := __Connect(username, password, url); err != nil {
		t.Fatal(err)
	} else {
		_gfile := __TestParseFile(t, __case_2)
		gfile := gofile(*_gfile)
		if _, err := gfile.Write(db); err != nil {
			t.Fatal(err)
		}
	}
}
