#!/bin/bash
clear
go run cli/main.go testrun 2>&1 | tee testrun.log
