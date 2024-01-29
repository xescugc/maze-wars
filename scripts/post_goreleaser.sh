#!/bin/bash

DIR=./dist

# For now when the bin is generated it's on a weird path
# due to goreleaser not supporting changes on the output.
# This script remove those weird outputs to the ones
# expected by goreleaser to continue
IFS='/'; binname=($1); unset IFS;
for d in "$1"/*; do
  if [ ! -d "$d" ]; then
    IFS='/'; paths=($d); unset IFS;
    cd $1
    mv ${paths[-1]} ../
    cd ../

    rm -rf ${binname[-1]}
    mv ${paths[-1]} ${binname[-1]}
  fi
done
