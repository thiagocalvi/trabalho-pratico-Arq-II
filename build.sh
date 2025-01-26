#!/bin/bash

BIN_DIR="executaveis"
FILE_NAME="trabalho-pratico-arq"

mkdir -p "$BIN_DIR"

GOOS=linux GOARCH=amd64 go build -o "$BIN_DIR/$FILE_NAME-linux-amd64" main.go memoria.go cache.go
GOOS=windows GOARCH=amd64 go build -o "$BIN_DIR/$FILE_NAME-windows-amd64.exe" main.go memoria.go cache.go