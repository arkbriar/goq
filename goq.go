package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"querygo"
	"querygo/debug"
	"strings"

	"time"
	"runtime"
)

const (
	username = "neo4j"
	password = "dsj1994"
	url      = "localhost:7474/db/data"
)

func ReadMemStats() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.TotalAlloc / 1024
}

func main() {
	log.SetOutput(os.Stdout)

	var SrcFile, ConfigFile, Neo4jConf string
	var Cypher, QueryParams string
	var DeleteItem string

	var DoList, Simple, DryRun bool

	var PrintVer bool

	var DebugLog debug.DebugLog

	flag.BoolVar(&PrintVer, "version", false, "print version")

	flag.StringVar(&SrcFile, "parse", "", "[project dir | go src file] parse dir/file, project name will automatically reducted")
	flag.StringVar(&ConfigFile, "config", "config.json", "[configuration file] use default config when not specified")
	flag.StringVar(&Neo4jConf, "neo4j", "neo4j:neo4j@localhost:7474", "[username:password@url:port] parameters for neo4j database, will prefer using what in config when config is specified")
	flag.StringVar(&Cypher, "query", "", "[cypher] do some built-in cypher queries, all patterns are: subprojects| packages_of_projects| structs_of_package| interfaces_of_package| interfaces_of_struct| structs_of_interface| inheritors_of_struct| structs_inherited_by  *OR*  subpros| pkgs_of_pro| stts_of_pkg| itfs_of_pkg| itfs_of_stt| stts_of_itf| ihts_of_stt| stts_ihted_by")

	flag.StringVar(&QueryParams, "param", "", "[query params] should use together with query option")
	flag.StringVar(&DeleteItem, "delete", "", "[project name:project | package name:package | file name:file] delete the named project(package, file) in neo4j")

	flag.BoolVar(&DoList, "list", false, "list the projects in neo4j")
	flag.BoolVar(&Simple, "simple", true, "delete any duplicated project when do a new export")
	flag.BoolVar(&DryRun, "dry-run", false, "do not export anything, just parse the project/file")

	flag.BoolVar((*bool)(&DebugLog), "debug", false, "show debug information")

	flag.Parse()

	if PrintVer {
		querygo.PrintVersion()
		os.Exit(0)
	}

	querygo.SetDebug(DebugLog)

	if err := querygo.ParseNeo4jConf(Neo4jConf); err != nil {
		log.Println(err)
	}

	if err := querygo.ParseConfFile(ConfigFile); err != nil {
		log.Println(err)
	}

	if !DryRun {
		querygo.Init()
	}

	if !DryRun && DoList {
		o, err := querygo.QueryProjects(querygo.DB)
		if err != nil {
			log.Fatalln(err)
		}
		if len(o) != 0 {
			fmt.Println("Projects in neo4j:")
			for _, i := range o {
				fmt.Println(i.First)
			}
		} else {
			fmt.Println("There're no projects in neo4j.")
		}
		os.Exit(0)
	}

	if SrcFile != "" {
		startUnixTime := time.Now().UnixNano()
		fmt.Printf("Memory allocated before parsing is %d Kb\n", ReadMemStats())
		if err := querygo.Export(SrcFile, DryRun, Simple); err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Memory allocated after parsing is %d Kb\n", ReadMemStats())
		stopUnixTime := time.Now().UnixNano()
		fmt.Printf("Time cost on parsing and exporting is %dms\n", (stopUnixTime - startUnixTime) / 1e6)
		os.Exit(0)
	}

	if !DryRun && Cypher != "" {
		Cypher = strings.ToUpper(Cypher)

		if QueryParams == "" {
			log.Fatalln("There must be a parammeter for querys.")
		}

		var err error
		var one []querygo.Oresult = nil
		var two []querygo.Tresult = nil
		var three []querygo.Thresult = nil

		var title string = "Result's are"

		/*
			subprojects, packages_of_projects, structs_of_package, interfaces_of_package, interfaces_of_struct, structs_of_interface, inheritors_of_struct, structs_inherited_by
			subpros, pkgs_of_pro, stts_of_pkg, itfs_of_pkg, itfs_of_stt, stts_of_itf, ihts_of_stt, stts_ihted_by
		*/

		switch Cypher {
		case "SUBPROJECTS":
			fallthrough
		case "SUBPROS":
			one, err = querygo.QuerySubProjects(querygo.DB, QueryParams)
			title = "Subprojects of " + QueryParams + " are"

		case "PACKAGES_OF_PROJECT":
			fallthrough
		case "PKGS_OF_PRO":
			one, err = querygo.QueryPackagesOfProject(querygo.DB, QueryParams)
			title = "Packages of project " + QueryParams + " are"

		case "STRUCTS_OF_PACKAGE":
			fallthrough
		case "STTS_OF_PKG":
			one, err = querygo.QueryStructsOfPackage(querygo.DB, QueryParams)
			title = "Structs of package " + QueryParams + " are"

		case "INTERFACES_OF_PACKAGE":
			fallthrough
		case "ITFS_OF_PKG":
			one, err = querygo.QueryInterfacesOfPackage(querygo.DB, QueryParams)
			title = "Interfaces of package " + QueryParams + " are"

		case "INTERFACES_OF_STRUCT":
			fallthrough
		case "ITFS_OF_STT":
			two, err = querygo.QueryInterfacesOfStruct(querygo.DB, QueryParams)
			title = "Interfaces of struct " + QueryParams + " are"

		case "STRUCTS_OF_INTERFACE":
			fallthrough
		case "STTS_OF_ITF":
			two, err = querygo.QueryStructsOfInterface(querygo.DB, QueryParams)
			title = "Structs of interface " + QueryParams + " are"

		case "INHERITORS_OF_STRUCT":
			fallthrough
		case "IHTS_OF_STT":
			two, err = querygo.QueryInheritorsOfStruct(querygo.DB, QueryParams)
			title = "Inheritors of struct " + QueryParams + " are"

		case "STRUCTS_INHERITED_BY":
			fallthrough
		case "STTS_IHTED_BY":
			two, err = querygo.QueryStructsInheritedBy(querygo.DB, QueryParams)
			title = "Structs inherited by " + QueryParams + " are"

		default:
			log.Fatalln("The query pattern is not defined.")
		}

		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(title)

		if one != nil {
			for _, x := range one {
				fmt.Println(x)
			}
		} else if two != nil {
			for _, x := range two {
				fmt.Println(x)
			}
		} else if three != nil {
			for _, x := range three {
				fmt.Println(x)
			}
		}

		os.Exit(0)
	}

	if !DryRun && DeleteItem != "" {
		ItemP := strings.Split(DeleteItem, ":")
		if len(ItemP) != 2 {
			log.Fatalln("Illegal form of delete item.")
		}

		ItemName, ItemLabel := ItemP[0], ItemP[1]

		ItemLabel = strings.ToLower(ItemLabel)

		var err error
		if ItemLabel == "project" {
			err = querygo.DeleteProject(querygo.DB, ItemName)
		} else if ItemLabel == "package" {
			err = querygo.DeletePackage(querygo.DB, ItemName)
		} else if ItemLabel == "file" {
			err = querygo.DeleteFile(querygo.DB, ItemName)
		} else {
			log.Fatalln("Illegal form of delete item.")
		}
		if err != nil {
			log.Fatalln(err)
		}

		os.Exit(0)
	}
}
