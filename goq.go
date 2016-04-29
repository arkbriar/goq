package main

import (
	"flag"
	"log"
	"os"
	"querygo"
	"fmt"
)

const (
	username = "neo4j"
	password = "dsj1994"
	url      = "localhost:7474/db/data"
)

func main() {
	log.SetOutput(os.Stdout)

	var SrcFile, ConfigFile, Neo4jConf string
	var Cypher, QueryParams string
	var DeleteItem string

	var DoList, Simple, DryRun bool

	var PrintVer bool

	var DebugLog querygo.DebugLog

	flag.BoolVar(&PrintVer, "version", false, "print version")

	flag.StringVar(&SrcFile, "parse", "", "[project dir | go src file] parse dir/file, project name will automatically reducted")
	flag.StringVar(&ConfigFile, "config", "config.json", "[configuration file] use default config when not specified")
	flag.StringVar(&Neo4jConf, "neo4j", "neo4j:neo4j@localhost:7474", "[username:password@url:port] parameters for neo4j database, will prefer using what in config when config is specified")
	flag.StringVar(&Cypher, "query", "", "[cypher] do some built-in cypher queries")
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

	querygo.Init()

	if DoList {
		o, err := querygo.QueryPackagesOfProject(querygo.DB, "golang")
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
		if err := querygo.Export(SrcFile, DryRun, Simple); err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)
	}

	if Cypher != "" {

		os.Exit(0)
	}

	if DeleteItem != "" {

		os.Exit(0)
	}
}
