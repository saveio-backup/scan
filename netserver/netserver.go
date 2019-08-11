package netserver

import (
	"fmt"
	"net"

	"github.com/saveio/scan/common/config"
	"github.com/saveio/scan/tracker"
	"github.com/saveio/themis/common/log"
)

type TKServer struct {
	Tsvr *tracker.Server
}

func NewTKServer() *TKServer {
	return &TKServer{
		Tsvr: tracker.NewServer(),
	}
}

func (ns *TKServer) Run() {
	go ns.StartTrackerListening()
}

func (ns *TKServer) StartTrackerListening() error {
	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", config.Parameters.Base.TrackerPortOffset))
	if err != nil {
		log.Errorf("start tracker service net.ListenPacket err:%s", err)
		return err
	}
	if pc != nil {
		defer pc.Close()
	}
	ns.Tsvr.SetPacketConn(pc)
	ns.Tsvr.Run()
	return nil
}
