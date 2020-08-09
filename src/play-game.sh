#!/bin/sh

# remove old files
#(cd client && rm client) || exit 1
#(cd server && rm server) || exit 1
#(cd broker && rm broker) || exit 1

# build
(cd client && go build) || exit 1
(cd server && go build) || exit 1
(cd broker && go build) || exit 1



#./server &
# give server time to create socket
#sleep 1
#
#./client &
#./client
#
#wait
#wait