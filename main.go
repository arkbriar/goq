package main

import (
	"fmt"
	"codelib"
	"os"
)

const (
	username = "neo4j"
	password = "dsj1994"
	url = "localhost:7474/db/data"
)

func main() {
	fmt.Println("Here is Code Library.")

	db, err:= codelib.ConnectToDB(username, password, url)
	if err != nil {
		fmt.Println(err.Error())
	}
	codelib.SetDB(db)

	err = codelib.ExportImportsInFileToDB(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
	}
}
