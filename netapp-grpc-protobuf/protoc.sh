#!/bin/bash
protoc -I keyval/ --go_out=plugins=grpc:keyval keyval/keyval.proto
