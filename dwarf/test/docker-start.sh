#!/bin/bash -e

# debugger need priviledges including, ptrace, etc.
count=`docker ps -a | grep debugger.env | wc -l` 

if [ $count -eq 0 ]
then
    docker run -it                                                              \
    -v `dirname $(dirname $(pwd -P))`:/root/dwarftest                           \
    --name dwarftest --cap-add ALL                                              \
    --rm dwarftest                                                              \
    /bin/bash
else
    docker exec -it dwarftest /bin/bash
fi
