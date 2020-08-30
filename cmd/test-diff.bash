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

duration=${1-"2"} # default 2
sleep $duration

for i in `ls . | grep -v .debug` ;
do
    echo -n "$i "
    # sync this 16 to ../helper.go#handshakeSize
    diff $i ${i:0:16}*.debug
    echo $?
done
