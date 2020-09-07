#!/usr/local/bin/bash

go build

# run server on another terminal
#./cmd -s

rm *.debug
for i in `ls .`;
do
    echo $i
    ./cmd $i &
done

duration=${1-"1"} # default 1
sleep $duration

for i in `ls . | grep -v .debug` ;
do
    echo -n "$i "
    diff $i ${i:0:16}*.debug
    echo $?
done
