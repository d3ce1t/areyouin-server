#!/usr/bin/env sh
go build -ldflags "-X main.BUILD_TIME=$(date -u '+%Y%m%d%.%H%M%S')" -o server.bin
