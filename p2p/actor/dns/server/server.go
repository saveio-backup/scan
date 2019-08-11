package server

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strconv"

	"github.com/saveio/scan/storage"

	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	p2pNet "github.com/saveio/carrier/network"
	pm "github.com/saveio/scan/messages/protoMessages"
	"github.com/saveio/scan/p2p/actor/messages"
	network "github.com/saveio/scan/p2p/networks/dns"
	"github.com/saveio/themis/common/log"
)

var P2pPid *actor.PID

type MessageHandler func(msgData interface{}, pid *actor.PID)

type P2PActor struct {
	net         *network.Network
	props       *actor.Props
	msgHandlers map[string]MessageHandler
	localPID    *actor.PID
}

func NewP2PActor() (*P2PActor, error) {
	var err error
	p2pActor := &P2PActor{
		msgHandlers: make(map[string]MessageHandler),
	}
	p2pActor.localPID, err = p2pActor.Start()
	if err != nil {
		return nil, err
	}
	return p2pActor, nil
}

func (this *P2PActor) SetNetwork(n *network.Network) {
	this.net = n
}

func (this *P2PActor) Start() (*actor.PID, error) {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	localPid, err := actor.SpawnNamed(this.props, "scan_dns_net_server")
	if err != nil {
		return nil, fmt.Errorf("[P2PActor] start error:%v", err)
	}
	this.localPID = localPid
	return localPid, err
}

func (this *P2PActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Restarting:
		log.Warn("[P2p]actor restarting")
	case *actor.Stopping:
		log.Warn("[P2p]actor stopping")
	case *actor.Stopped:
		log.Warn("[P2p]actor stopped")
	case *actor.Started:
		log.Debug("[P2p]actor started")
	case *actor.Restart:
		log.Warn("[P2p]actor restart")
	case *pm.Registry:
		log.Debugf("tracker client registry or update msg:%s", msg.String())
		if msg.WalletAddr == "" || msg.HostPort == "" {
			log.Errorf("[MSB Receive] receive from peer:%s, nil Reg message", ctx.Sender().Address)
			break
		}
		nodeAddr, err := storage.EDB.GetEndpoint(msg.WalletAddr)
		if err != nil {
			return
		}

		if msg.Type == 0 {
			msg.Type = 1
			// go this.Broadcast(msg)
		} else if msg.Type == 1 && fmt.Sprintf("%s", nodeAddr.NodeAddr) != msg.HostPort {
			err := UpdateEndpointDB(msg.WalletAddr, msg.HostPort)
			if err != nil {
				log.Errorf("[MSB Receive] sync registry failed, error:%v", err)
			}
			log.Debugf("[MSB Receive] sync registry success: %v", msg.String())
		}
	case *pm.UnRegistry:
		log.Debugf("tracker client unRegistry msg:%s", msg.String())
		if msg.WalletAddr == "" {
			log.Errorf("[MSB Receive] receive from peer:%s, nil Reg message", ctx.Sender().Address)
			break
		}

		nodeAddr, err := storage.EDB.GetEndpoint(msg.WalletAddr)
		if err != nil {
			return
		}

		if msg.Type == 0 {
			msg.Type = 1
			go this.Broadcast(msg)
		} else if msg.Type == 1 && fmt.Sprintf("%s", nodeAddr.NodeAddr) != "" {
			err := storage.EDB.DelEndpoint(msg.WalletAddr)
			if err != nil {
				log.Errorf("[MSB Receive] sync regmessage error:%v", err)
			}
		}
		// go this.Broadcast(msg)
	case *pm.Torrent:
		log.Debugf("tracker client torrent msg:%s", msg.String())
		if msg.InfoHash == nil || msg.Torrent == nil {
			log.Errorf("[MSB Receive] receive from peer:%s, nil Torrent message", ctx.Sender().Address)
			break
		}

		hpBytes, _ := storage.TDB.Get(msg.InfoHash)
		if msg.Type == 0 {
			msg.Type = 1
			go this.Broadcast(msg)
		} else if msg.Type == 1 && string(hpBytes) != string(msg.Torrent) {
			err := storage.TDB.Put(msg.InfoHash, msg.Torrent)
			if err != nil {
				log.Errorf("[MSB Receive] sync filemessage error:%v", err)
			}
		} else {
			log.Errorf("No NEED BROADCAST, hpBytes: %s, v: %s", hpBytes, msg.Torrent)
		}
	case *messages.UserDefineMsg:
		t := reflect.TypeOf(msg)
		this.msgHandlers[t.Name()](msg, this.localPID)
	default:
		log.Error("[P2PActor] receive unknown message type!")
	}

}

func (this *P2PActor) Broadcast(message proto.Message) {
	ctx := p2pNet.WithSignMessage(context.Background(), true)
	this.net.P2p.Broadcast(ctx, message)
}

func (this *P2PActor) RegMsgHandler(msgName string, handler MessageHandler) {
	this.msgHandlers[msgName] = handler
}

func (this *P2PActor) UnRegMsgHandler(msgName string, handler MessageHandler) {
	delete(this.msgHandlers, msgName)
}

func (this *P2PActor) GetLocalPID() *actor.PID {
	return this.localPID
}

func UpdateEndpointDB(walletAddr, hostAddr string) error {
	host, port, err := net.SplitHostPort(hostAddr)
	if err != nil {
		return err
	}
	netIp := net.ParseIP(host).To4()
	if netIp == nil {
		netIp = net.ParseIP(host).To16()
	}
	netPort, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	return storage.EDB.PutEndpoint(walletAddr, netIp, netPort)
}
