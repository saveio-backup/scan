package netserver

import (
	"fmt"
	"net"

	"github.com/saveio/scan/common/config"
	"github.com/saveio/scan/dns"
	"github.com/saveio/scan/tracker"
	"github.com/saveio/themis/common/log"
)

var (
	TRACKER_DB_PATH = "./torrentdb"
)

type NetServer struct {
	Tsvr *tracker.Server
	dsvr *dns.Server
}

func NewNetServer() *NetServer {
	return &NetServer{
		Tsvr: tracker.NewServer(),
		dsvr: dns.NewServer(),
	}
}

// Start start netserver service
func (ns *NetServer) Run() error {
	go ns.startTrackerListening()
	//go ns.startSyncNet()
	return nil
}

// startTrackerListening start tracker listen
func (ns *NetServer) startTrackerListening() error {
	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", config.Parameters.Base.TrackerPortOffset))
	if err != nil {
		log.Errorf("start tracker service net.ListenPacket err:%s", err)
		return err
	}
	if pc != nil {
		defer pc.Close()
	}
	ns.Tsvr.SetPacketConn(pc)
	err = ns.Tsvr.Run()
	if err != nil {
		log.Errorf("start tracker service err:%s", err)
	}
	return err
}
