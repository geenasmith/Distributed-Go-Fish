package common

import "strconv"
import "os"

// RPC's:
// - start a game
// - Ask for a card
// - play a pair
// - game status (in progress, your turn, game over)


// ** adapted from mapreduce
func GameServerSock() string {
	s := "/var/tmp/824-gs-"
	s += strconv.Itoa(os.Getuid())
	return s
}