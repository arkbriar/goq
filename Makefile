.PHONY: binary default test clean

BASH=bash

default: binary

binary:
	@ $(BASH) hack/make.sh BINARY

test:
	@ $(BASH) hack/test.sh

clean:
	- rm main
