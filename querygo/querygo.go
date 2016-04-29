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

func ExportProject(dir string, dryrun bool, simple bool) error {
	_gpro, err := golang.ParseProject(dir)
	if err != nil {
		return err
	}

	if !dryrun {
		gpro := gopro(*_gpro)

		if simple {
			if _err := DeleteProject(DB, _gpro.Name); _err != nil {
				return _err
			}
		}

		_, err = gpro.Write(DB)
		if err != nil {
			if _err := DeleteProject(DB, _gpro.Name); _err != nil {
				return _err
			}

			return err
		}
	}

	return nil
}

func ExportFile(file *os.File, dryrun bool, simple bool) error {
	_gfile, err := golang.ParseFile(file)
	if err != nil {
		return err
	}

	if !dryrun {
		gfile := gofile(*_gfile)

		if simple {
			if _err := DeleteFile(DB, _gfile.Name); _err != nil {
				return _err
			}
		}

		_, err = gfile.Write(DB)
		if err != nil {
			if _err := DeleteFile(DB, _gfile.Name); _err != nil {
				return _err
			}

			return err
		}
	}

	return nil
}

func Export(file string, dryrun bool, simple bool) error {
	if f, err := os.Open(file); err != nil {
		return err
	} else {
		fstat, err := f.Stat()
		if err != nil {
			return err
		}
		if fstat.IsDir() {
			return ExportProject(file, dryrun, simple)
		} else {
			return ExportFile(f, dryrun, simple)
		}
	}

	return nil
}

func ExportEveryProjectInGoPath() error {
	return ExportProject(GOPATH, false, true)
}

func PrintVersion() {
	fmt.Println("QueryGo Version 1.0")
}
