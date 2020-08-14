#!/bin/bash
status=0
lines=$(cat $1 | wc -l)
for f in $(find . -name "*.go")
do
    diff=$(head $f -n $lines | diff $1 -)
    if [ ! -z "$diff" ]
    then
        echo $f
        echo "< want, > got"
        echo "${diff}"
        echo ""
        status=1
    fi
done
exit $status
