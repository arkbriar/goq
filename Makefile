.PHONY: binary default test clean

BASH=bash

default: binary

binary:
	@ $(BASH) hack/make.sh BINARY

test: build
	hack/test.sh

clean:
	- rm main
