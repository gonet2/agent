#!/bin/bash -e

protoc  ./*.proto --go_out=../protos
