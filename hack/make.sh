#!/bin/bash

function Build() {
    echo "Building ..."
    go build -o $1 .
}

Build $1;
