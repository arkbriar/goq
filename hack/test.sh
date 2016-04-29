#!/bin/bash

function Test() {
    go test -cover -v .
}

pushd querygo

# test package codelib/golang
pushd golang
Test;
popd

# test package codelib
Test;

popd
