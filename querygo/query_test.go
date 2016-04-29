// Copyright 2016 ArkBriar. All rights reserved.
package querygo

import (
	"querygo/golang"
	"testing"

	"github.com/jmcvetta/neoism"
)

var query_test_db *neoism.Database
var query_test_pro *golang.GoProject

// using consts and functions in neo4j_test.go
func __Parse_Export(t *testing.T, dir string) {
	if db, err := __Connect(username, password, url); err != nil {
		t.Fatal(err)
	} else {
		query_test_db = db
		t.Log("Parsing " + dir + "...")
		if _gpro, err := golang.ParseProject(dir); err != nil {
			t.Fatal(err)
		} else {
			query_test_pro = _gpro
			gpro := gopro(*_gpro)
			t.Log("Writing to neo4j...")
			if _, err := gpro.Write(db); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func __Delete(t *testing.T, db *neoism.Database, name string) {
	// then delete
	t.Log("Deleting...")
	if err := DeleteProject(db, name); err != nil {
		t.Fatal(err)
	}
}

func __Query(t *testing.T, db *neoism.Database, query interface{}, name string) {
	switch query.(type) {
	case func(*neoism.Database, string)([]Oresult, error):
		if r, err := (query.(func(*neoism.Database, string)([]Oresult, error)))(db, name); err != nil {
			t.Fatal(err)
		} else {
			t.Log(r)
		}
	case func(*neoism.Database, string)([]Tresult, error):
		if r, err := (query.(func(*neoism.Database, string)([]Tresult, error)))(db, name); err != nil {
			t.Fatal(err)
		} else {
			t.Log(r)
		}
	case func(*neoism.Database, string)([]Thresult, error):
		if r, err := (query.(func(*neoism.Database, string)([]Thresult, error)))(db, name); err != nil {
			t.Fatal(err)
		} else {
			t.Log(r)
		}
	default:
		// should never reach here
		t.Fatal("query function illegal!")
	}
}

// first
func TestParse(t *testing.T) {
	// self parse
	__Parse_Export(t, ".")
}

func TestQueryInheritorsOfStruct(t *testing.T) {
	__Query(t, query_test_db, QueryInheritorsOfStruct, "GoFunc")
}

func TestQueryInterfacesOfPackage(t *testing.T) {
	__Query(t, query_test_db, QueryInterfacesOfPackage, "querygo")
}

func TestQueryInterfacesOfStruct(t *testing.T) {
	__Query(t, query_test_db, QueryInterfacesOfStruct, "gopro")
}

func TestQueryPackagesOfProject(t *testing.T) {
	__Query(t, query_test_db, QueryPackagesOfProject, "querygo")
}

func TestQueryProjects(t *testing.T) {
	if r, err := QueryProjects(query_test_db); err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
}

func TestQueryStructsInheritedBy(t *testing.T) {
	__Query(t, query_test_db, QueryStructsInheritedBy, "GoMethod")
}

func TestQueryStructsOfInterface(t *testing.T) {
	__Query(t, query_test_db, QueryStructsOfInterface, "Neo4jMap")
}

func TestQueryStructsOfPackage(t *testing.T) {
	__Query(t, query_test_db, QueryStructsOfPackage, "golang")
}

func TestQuerySubProjects(t *testing.T) {
	__Query(t, query_test_db, QuerySubProjects, "querygo")
}

func TestDelete(t *testing.T) {
	__Delete(t, query_test_db, query_test_pro.Name)
}
