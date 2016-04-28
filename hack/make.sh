#!/bin/bash

function Build() {
    echo "Building ..."
    go install codelib
    go build -o GoQuery main.go
}

if [ $# == 0 ]; then
    exit -1
fi

if [ $1 = "BINARY" ]; then
    Build;
fi
