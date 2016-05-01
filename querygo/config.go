// Copyright 2016 ArkBriar. All rights reserved.
package querygo

import (
	"errors"
	"strconv"
	"strings"
)

var (
	Username string = "neo4j"
	Password string = "neo4j"
	NeoUrl   string = "localhost"
	Port     int    = 7474
)

func ParseNeo4jConf(conf string) error {
	u := strings.Split(conf, "@")
	if len(u) != 2 {
		return errors.New("invalid configuration for neo4j")
	}

	upa := strings.Split(u[0], ":")
	upo := strings.Split(u[1], ":")

	if len(upa) != 2 || len(upo) != 2 {
		return errors.New("invalid configuration for neo4j")
	}

	Username = upa[0]
	Password = upa[1]
	NeoUrl = upo[0]
	Port, _ = strconv.Atoi(upo[1])

	return nil
}

func ParseConfFile(file string) error {
	//@TODO
	return nil
}

