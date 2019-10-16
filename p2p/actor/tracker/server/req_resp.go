package server

import (
	pm "github.com/saveio/scan/p2p/actor/messages"
)

type AnnounceResponse struct {
}

type AnnounceRet struct {
	Ret  *pm.AnnounceResponse
	Done chan bool
	Err  error
}

type AnnounceReq struct {
	Announce *pm.AnnounceRequest
	Ret      *AnnounceRet
}
