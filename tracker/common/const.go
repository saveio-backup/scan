package common

const (
	SYNC_MSG_OP_CODE      = 3000
	SYNC_REGMSG_OP_CODE   = 3001
	SYNC_UNREGMSG_OP_CODE = 3002
)

var ListeningCh = make(chan struct{})
