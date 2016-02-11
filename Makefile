.PHONY: binary build default deb rpm test

default: binary

binary: build
	hack/make.sh binary

build:
	hack/make.sh

deb: build
	hack/make.sh deb

rpm: build
	hack/make.sh rpm

test: build
	hack/test.sh
