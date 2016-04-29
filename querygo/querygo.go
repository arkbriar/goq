// Copyright 2016 ArkBriar. All rights reserved.
package querygo

import (
	"fmt"
	"os"
	"querygo/golang"
	"strconv"
)

func Init() {
	var err error
	if DB, err = ConnectNeo4j(
		Username,
		Password,
		NeoUrl+":"+strconv.Itoa(Port)+"/db/data",
	); err != nil {
		fmt.Fprintln(os.Stderr, "Can not connect to neo4j database: "+
			"http://"+Username+":"+Password+"@"+NeoUrl+":"+strconv.Itoa(Port),
			"Please check your configuration")
		os.Exit(-1)
	}
}

func ExportProject(dir string) error {
	_gpro, err := golang.ParseProject(dir)
	if err != nil {
		return err
	}

	gpro := gopro(*_gpro)
	_, err = gpro.Write(DB)
	if err != nil {
		if _err := DeleteProject(DB, _gpro.Name); _err != nil {
			return _err
		}

		return err
	}

	return nil
}

func ExportEveryProjectInGoPath() error {
	return ExportProject(GOPATH)
}

func PrintVersion() {

}
