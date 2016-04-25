// Copyright 2016 ArkBriar. All rights reserved.
package codelib

import (
	"codelib/golang"
	"os"
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
	pkgcase_2 = "ast"
)

func __TestParseFile(t *testing.T, file string) *golang.GoFile {
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

/*
 *func TestExportGoFileToNeo4j(t *testing.T) {
 *    if db, err := __Connect(username, password, url); err != nil {
 *        t.Fatal(err)
 *    } else {
 *        _gfile := __TestParseFile(t, filecase_1)
 *        gfile := gofile(*_gfile)
 *        if _, err := gfile.Write(db); err != nil {
 *            t.Fatal(err)
 *        }
 *    }
 *}
 *
 *func TestExportGoFileToNeo4j2(t *testing.T) {
 *    if db, err := __Connect(username, password, url); err != nil {
 *        t.Fatal(err)
 *    } else {
 *        _gfile := __TestParseFile(t, filecase_2)
 *        gfile := gofile(*_gfile)
 *        if _, err := gfile.Write(db); err != nil {
 *            t.Fatal(err)
 *        }
 *    }
 *}
 */

func TestExportGoPackageToNeo4j1(t *testing.T) {
	if db, err := __Connect(username, password, url); err != nil {
		t.Fatal(err)
	} else {
		if _gpro, err := golang.ParseProject(testdir + "/" + pkgcase_1); err != nil {
			t.Fatal(err)
		} else {
			gpro := gopro(*_gpro)
			if _, err := gpro.Write(db); err != nil {
				t.Fatal(err)
			}
		}
	}
}

/*
 *func TestExportGoPackageToNeo4j2(t *testing.T) {
 *    if db, err := __Connect(username, password, url); err != nil {
 *        t.Fatal(err)
 *    } else {
 *        if _gpro, err := golang.ParseProject(testdir + "/" + pkgcase_2); err != nil {
 *            t.Fatal(err)
 *        } else {
 *            gpro := gopro(*_gpro)
 *            if _, err := gpro.Write(db); err != nil {
 *                t.Fatal(err)
 *            }
 *        }
 *    }
 *}
 */
