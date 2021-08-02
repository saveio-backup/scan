package common

import "os"

// network common
const (
	ACK_MSG_CHECK_INTERVAL = 20 // ack msg check interval
	MAX_ACK_MSG_TIMEOUT    = 60 // max timeout for ack msg
)

// FileExisted checks whether filename exists in filesystem
func FileExisted(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
