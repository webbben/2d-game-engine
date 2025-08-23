#!/bin/bash

if [ -f testrun.log ]; then
    rm testrun.log
fi

clear
go run cli/main.go testrun 2>&1 | tee testrun.log
