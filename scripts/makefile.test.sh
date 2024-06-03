#!/bin/bash
# file: makefile.test.sh
# url: https://github.com/conneroisu/seltabl/scripts/makefile.test.sh
# title: Test Script
# description: This script runs the test for the project.
# 
# usage: make test

gum spin --spinner dot --title "Running Go Test With Race" --show-output -- \
    go test -race -v -timeout 30s ./...
# gum spin --spinner dot --title "Running Go Test With Coverage" --show-output -- \
go test -coverprofile=coverage.out ./... 
# gum spin --spinner dot --title "Running Make Lint" --show-output -- \
    # make lint
    #

# if gocovsh is executable
if [ -x "$(command -v gocovsh)" ]; then
    # if gocovsh is not empty
    if [ -s coverage.out ]; then
        # run gocovsh
        gocovsh
    else
        # if gocovsh is empty
        echo "No coverage.out file found."
    fi
else
    # if gocovsh is not executable
    echo "gocovsh is not executable."
fi