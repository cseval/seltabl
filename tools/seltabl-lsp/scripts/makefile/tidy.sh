#!/bin/bash
# file: makefile.tidy.sh
# url: https://github.com/conneroisu/seltab/tools/seltab-lsp/scripts/makefile.tidy.sh
# title: Running Go Mod Tidy
# description: This script runs go mod tidy to clean up the go.mod and go.sum files.
#
# Usage: make tidy

go mod tidy
