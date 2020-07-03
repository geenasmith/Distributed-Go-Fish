package common

import "strconv"
import "os"

// ** adapted from mapreduce
func GameServerSock() string {
	s := "/var/tmp/824-gs-"
	s += strconv.Itoa(os.Getuid())
	return s
}
