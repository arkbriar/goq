// Copyright 2016 ArkBriar. All rights reserved.
package querygo

import (
	"os"
	"querygo/golang"
	"testing"
)

const (
	username = "neo4j"
	password = "dsj1994"
	url      = "localhost:7474/db/data"
)

const (
	testdir = "golang/testcases"

	filecase_1 = "types.go"
	filecase_2 = "ast.go"
	pkgcase_1  = "pkgtest"
	pkgcase_2  = "ast"
)

func __TestParseFile(t *testing.T, file string) *golang.GoFile {
	if file, err := os.Open(file); err != nil {
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

func __Parse_Export_Then_Delete_File(t *testing.T, file string) {
	if db, err := __Connect(username, password, url); err != nil {
		t.Fatal(err)
	} else {
		t.Log("Parsing " + filecase_1 + "...")
		_gfile := __TestParseFile(t, file)
		gfile := gofile(*_gfile)
		t.Log("Writing to neo4j...")
		if _, err := gfile.Write(db); err != nil {
			t.Fatal(err)
		}

		// then delete
		t.Log("Deleting...")
		if err := DeleteFile(db, gfile.Name); err != nil {
			t.Fatal(err)
		}
	}
}

func __Parse_Export_Then_Delete_Project(t *testing.T, dir string) {
	if db, err := __Connect(username, password, url); err != nil {
		t.Fatal(err)
	} else {
		t.Log("Parsing " + dir + "...")
		if _gpro, err := golang.ParseProject(dir); err != nil {
			t.Fatal(err)
		} else {
			gpro := gopro(*_gpro)
			t.Log("Writing to neo4j...")
			if _, err := gpro.Write(db); err != nil {
				t.Fatal(err)
			}

			// then delete
			t.Log("Deleting...")
			if err := DeleteProject(db, gpro.Name); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestFileParseAndExport(t *testing.T) {
	__Parse_Export_Then_Delete_File(t, testdir+"/"+filecase_1)
	__Parse_Export_Then_Delete_File(t, testdir+"/"+filecase_2)
}

func TestProjectParseAndExport(t *testing.T) {
	__Parse_Export_Then_Delete_Project(t, testdir+"/"+pkgcase_1)
	__Parse_Export_Then_Delete_Project(t, testdir+"/"+pkgcase_2)
}

func TestSelfParseAndExport(t *testing.T) {
	__Parse_Export_Then_Delete_Project(t, ".")
}
