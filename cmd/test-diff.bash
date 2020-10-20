#!/usr/local/bin/bash

go build

# make 5.gb for cpu profile
# dd if=/dev/zero of=5gb bs=1024 count=5000000

# run server on another terminal
#./cmd -s

rm *.debug
for i in `ls .`;
do
    if [ -f `pwd`/$i ]
    then
        echo $i
        ./cmd $i &
    else
        echo dir $i
    fi
done

duration=${1-"1"} # default 1
sleep $duration

for i in `ls . | grep -v .debug` ;
do
    if [ -f `pwd`/$i ]
    then
        echo -n "$i "
        diff $i ${i:0:16}*.debug
        echo $?
    else
        echo dir $i again
    fi
done
