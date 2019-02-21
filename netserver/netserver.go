package netserver

import (
	"fmt"
	"net"

	"github.com/oniio/oniChain/common/log"
	"github.com/oniio/oniDNS/config"
	"github.com/oniio/oniDNS/dns"
	"github.com/oniio/oniDNS/tracker"
)
var (
	TRACKER_DB_PATH="./torrentdb"
)

type NetServer struct {
	tsvr *tracker.Server
	dsvr *dns.Server
}

func NewNetServer() *NetServer {
	return &NetServer{
		tsvr: tracker.NewServer(TRACKER_DB_PATH),
		dsvr: dns.NewServer(),
	}
}

// Start start netserver service
func (ns *NetServer) Run() error {
	go ns.startTrackerListening()
	return nil
}

// startTrackerListening start tracker listen
func (ns *NetServer) startTrackerListening() error {
	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", config.DefaultConfig.Tracker.UdpPort))
	defer pc.Close()
	if err != nil {
		return err
	}
	ns.tsvr.SetPacketConn(pc)
	err = ns.tsvr.Run()
	if err != nil {
		log.Errorf("start tracker service err:%s", err)
	}
	return err
}

// startDnsListening start dns listen
func (ns *NetServer) startDnsListening() error {
	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", config.DefaultConfig.Dns.UdpPort))
	defer pc.Close()
	if err != nil {
		return err
	}
	return nil
}
