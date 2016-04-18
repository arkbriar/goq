package golang

import (
	"testing"
	"os"
	"codelib/golang"
)

const (
	testdir = "testcases"
)

func __TestParseFile(t *testing.T, file string) {
	if file, err := os.Open(testdir + "/" + file); err != nil {
		t.Fatal(err)
	} else {
		defer file.Close()

		_, err := golang.ParseFile(file)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestParseFile(t *testing.T) {
	__TestParseFile(t, "types.go")
}

func TestParseFile2(t *testing.T) {
	__TestParseFile(t, "ast.go")
}
