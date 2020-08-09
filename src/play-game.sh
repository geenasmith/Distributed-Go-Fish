#!/bin/sh

# build
(cd client && go build) || exit 1
(cd server && go build) || exit 1
(cd broker && go build) || exit 1
rm -f server-id || exit 1

# start the broker
echo '***' Starting the broker '***'
./broker/broker &
sleep 1

# start the server
echo '***' Starting the server '***'
./server/server &
sleep 1

# create 2 players
echo '***' creating 2 players '***'
./client/client &
./client/client

wait
wait
wait

#./server &
# give server time to create socket
#sleep 1
#
#./client &
#./client
#
#wait
#wait