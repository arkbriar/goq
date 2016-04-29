#!/bin/bash

function Build() {
    echo "Building ..."
    go install querygo
    go build -o $1 .
}

Build $1;
