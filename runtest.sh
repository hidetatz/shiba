#!/bin/bash

./task goimports
./task build

for fname in ./tests/*.sh; do
        ./"$fname"
done
