#!/bin/bash

function Test() {
    echo "Running all tests of" $(pwd) "..."
    go test -v .
}

pushd codelib

# test package codelib/golang
pushd golang
Test;
popd

# test package codelib
Test;

popd
