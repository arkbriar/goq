// Copyright 2016 ArkBriar. All rights reserved.
package codelib

import (
	"github.com/jmcvetta/neoism"
)

// simple queries
const (
	__QUERY_PROJECTS string = `
	MATCH (x:PROJECT)
	RETURN x.name
	`
	__QUERY_SUBPROJECTS string = `
	MATCH (p:PROJECT)-[:HAS]-[x:PROJECT]
	WHERE p.name = {name}
	RETURN x.name
	`

	__QUERY_PACKAGES_OF_PROJECT string = `
	MATCH (x:PACKAGE)<-[:HAS]-[p:PROJECT]
	WHERE p.name = {name}
	RETURN x.name
	`

	__QUERY_STRUCTS_OF_PACKAGE string = `
	MATCH (x:STRUCT)<-[:DEFINE]-[p:PACKAGE]
	WHERE p.name = {name}
	RETURN x.name
	`

	__QUERY_INTERFACES_OF_PACKAGE string = `
	MATCH (x:INTERFACE)<-[:DEFINE]-[p:PACKAGE]
	WHERE p.name = {name}
	RETURN x.name
	`
	// `struct` can be struct or alias
	// properties are in something like `{name}`
	__QUERY_INTERFACES_OF_STRUCT string = `
	MATCH (s:TYPE)-[:IMPLEMENT]-(x:INTERFACE)-[:DEFINE]-(f:FILE)
	WHERE s.name = {name}
	RETURN x.name, y.name

	`
	__QUERY_STRUCTS_OF_INTERFACE string = `
	MATCH (f:FILE)-[:DEFINE]-(x:TYPE)-[:IMPLEMENT]-(i:INTERFACE)
	WHERE i.name = {name}
	RETURN x.name, y.name
	`

	__QUERY_INHERITORS_OF_STRUCT string = `
	MATCH (f:FILE)-[:DEFINE]-(x:TYPE)-[:EXTEND]-(t:TYPE)
	WHERE t.name = {name}
	RETURN x.name, y.name
	`

	__QUERY_STRUCTS_INHERITED_BY string = `
	MATCH (s:TYPE)-[:EXTEND]-(x:TYPE)-[:DEFINE]-(f:FILE)
	WHERE s.name = {name}
	RETURN x.name, y.name
	`
)

// when res == nil, the res in cyperquery should be preset, or it will return an error
func query(db *neoism.Database, cq *neoism.CypherQuery, res interface{}) error {
	// set the result field
	if res != nil {
		cq.Result = res
	}

	// do the cypher query
	if err := db.Cypher(cq); err != nil {
		return err
	}

	return nil
}

type Oresult struct {
	first string `json:"x.name"`
}

type Tresult struct {
	first  string `json:"x.name"`
	second string `json:"y.name"`
}

type Thresult struct {
	first  string `json:"x.name"`
	second string `json:"y.name"`
	third  string `json:"z.name"`
}

func CreateCypherQuery(stmt string, params map[string]interface{}, res interface{}) *neoism.CypherQuery {
	return &neoism.CypherQuery{
		Statement:  stmt,
		Parameters: params,
		Result:     res,
	}
}

func internalImplementationOfSimpleQuery1(db *neoism.Database, QUERY string, name string) ([]Oresult, error) {
	res := make([]Oresult, 0, 4)
	if err := query(
		db,
		CreateCypherQuery(
			QUERY,
			neoism.Props{"name": name},
			res,
		),
		nil,
	); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func QueryProjects(db *neoism.Database) ([]Oresult, error) {
	res := make([]Oresult, 0, 4)
	if err := query(
		db,
		CreateCypherQuery(
			__QUERY_PROJECTS,
			nil,
			res,
		),
		nil,
	); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func QuerySubProjects(db *neoism.Database, name string) ([]Oresult, error) {
	return internalImplementationOfSimpleQuery1(db, __QUERY_SUBPROJECTS, name)
}

func QueryPackagesOfProject(db *neoism.Database, name string) ([]Oresult, error) {
	return internalImplementationOfSimpleQuery1(db, __QUERY_PACKAGES_OF_PROJECT, name)
}

func QueryStructsOfPackage(db *neoism.Database, name string) ([]Oresult, error) {
	return internalImplementationOfSimpleQuery1(db, __QUERY_STRUCTS_OF_PACKAGE, name)
}

func QueryInterfacesOfPackage(db *neoism.Database, name string) ([]Oresult, error) {
	return internalImplementationOfSimpleQuery1(db, __QUERY_INTERFACES_OF_PACKAGE, name)
}

func internalImplementationOfSimpleQuery2(db *neoism.Database, QUERY string, name string) ([]Tresult, error) {
	res := make([]Tresult, 0, 4)
	if err := query(
		db,
		CreateCypherQuery(
			QUERY,
			neoism.Props{"name": name},
			res,
		),
		nil,
	); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func QueryInterfacesOfStruct(db *neoism.Database, name string) ([]Tresult, error) {
	return internalImplementationOfSimpleQuery2(db, __QUERY_INTERFACES_OF_STRUCT, name)
}

func QueryStructsOfInterface(db *neoism.Database, name string) ([]Tresult, error) {
	return internalImplementationOfSimpleQuery2(db, __QUERY_STRUCTS_OF_INTERFACE, name)
}

func QueryInheritorsOfStruct(db *neoism.Database, name string) ([]Tresult, error) {
	return internalImplementationOfSimpleQuery2(db, __QUERY_INHERITORS_OF_STRUCT, name)
}

func QueryStructsInheritedBy(db *neoism.Database, name string) ([]Tresult, error) {
	return internalImplementationOfSimpleQuery2(db, __QUERY_STRUCTS_INHERITED_BY, name)
}
