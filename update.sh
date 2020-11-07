#!/bin/sh -e
# This sample script creates (or updates) checksum file
# with checksums of all files under lib/
go build
./checksum $@ test.sum lib
