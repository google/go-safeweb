#!/bin/bash
status=0
lines=$(cat $1 | wc -l)
extension=$2
for f in $(find . -name "*.$extension")
do
    diff=$(head -n $lines $f | sed 's/20[0-9][0-9]/YEAR/g' | diff $1 -)
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
