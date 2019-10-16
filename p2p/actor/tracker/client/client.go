package client

import (
	"net"

	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/scan/storage"
	"github.com/saveio/themis/common/log"
)

const (
	REQ_TIMEOUT = 15
)

var TrackerServerPid *actor.PID

func SetTrackerServerPid(p2pPid *actor.PID) {
	TrackerServerPid = p2pPid
}

type ConnectRet struct {
	Done chan bool
	Err  error
}

type ConnectReq struct {
	Address string
	Ret     *ConnectRet
}

type CloseRet struct {
	Done chan bool
	Err  error
}

type CloseReq struct {
	Address string
	Ret     *CloseRet
}

type SendRet struct {
	Done chan bool
	Err  error
}

type SendReq struct {
	Address string
	Data    proto.Message
	Ret     *SendRet
}

type RecvMsgRet struct {
	Done chan bool
	Err  error
}

type RecvMsg struct {
	From    string
	Message proto.Message
	Ret     *RecvMsgRet
}

func P2pConnect(address string) error {
	log.Infof("p2pConnnect address: %s", address)
	ret := &ConnectRet{
		Done: make(chan bool, 1),
		Err:  nil,
	}
	conRet := &ConnectReq{Address: address, Ret: ret}
	TrackerServerPid.Tell(conRet)
	<-conRet.Ret.Done
	close(conRet.Ret.Done)
	return conRet.Ret.Err
}

func P2pClose(address string) error {
	ret := &CloseRet{
		Done: make(chan bool, 1),
		Err:  nil,
	}
	chReq := &CloseReq{Address: address, Ret: ret}
	TrackerServerPid.Tell(chReq)
	<-chReq.Ret.Done
	close(chReq.Ret.Done)
	return chReq.Ret.Err
}

func P2pSend(address string, data proto.Message) error {
	log.Infof("p2pSend address: %s, data: %v", address, data)
	ret := &SendRet{
		Done: make(chan bool, 1),
		Err:  nil,
	}
	chReq := &SendReq{Address: address, Data: data, Ret: ret}
	TrackerServerPid.Tell(chReq)
	<-chReq.Ret.Done
	close(chReq.Ret.Done)
	return chReq.Ret.Err
}

func P2pTell() {
	log.Infof("P2pTell")
	TrackerServerPid.Tell(&actor.Started{})
	TrackerServerPid.Tell(&actor.Started{})
}

type ActEndpointRegistParams struct {
	Wallet [20]byte
	Ip     net.IP
	Port   uint64
}

type ActGetEndpointAddrParams struct {
	Wallet [20]byte
}

type ActCompleteTorrentParams struct {
	InfoHash storage.MetaInfoHash
	Ip       net.IP
	Port     uint64
}

type ActTorrentPeersParams struct {
	InfoHash storage.MetaInfoHash
	NumWant  int32
	Left     uint64
}

type ActNodesTypeRegistParams struct {
}

type ActGetNodesByTypeParams struct {
}
