#!/bin/sh -e
go build
./checksum $@ test.sum lib
