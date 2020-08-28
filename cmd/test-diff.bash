#!/usr/local/bin/bash

go build

# donot run this on background
#./cmd -s &

rm *.debug
for i in `ls .`;
do
    echo $i
    ./cmd $i &
done

time=${1-"3"} # default 3
sleep $time

for i in `ls . | grep -v .debug` ;
do
    echo -n "$i "
    diff $i ${i:0:8}*.debug
    echo $?
done
