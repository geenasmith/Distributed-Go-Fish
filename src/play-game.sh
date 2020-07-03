#!/bin/sh

# build
go build client.go || exit 1
go build server.go || exit 1

./server &
# give server time to create socket
sleep 1

./client &
./client

wait
wait