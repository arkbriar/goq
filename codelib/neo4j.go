// Copyright 2016 ArkBriar. All rights reserved.
package codelib

import (
	"github.com/jmcvetta/neoism"
)

func ConnectToDB(username, password, url string) (db *neoism.Database, err error) {
	return neoism.Connect("http://" + username + ":" + password + "@" + url);
}

