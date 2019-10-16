package server

import (
	"context"
	"fmt"

	"github.com/saveio/scan/storage"

	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	p2pNet "github.com/saveio/carrier/network"
	pm "github.com/saveio/scan/p2p/actor/messages"
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

func (this *P2PActor) GetNetwork() *network.Network {
	return this.net
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
	case *pm.Endpoint:
		log.Debugf("dns actor receive pm.Registry msg, WalletAddr: %s, HostPort: %s\n", msg.WalletAddr, msg.HostPort)
		if msg.WalletAddr == "" || msg.HostPort == "" {
			break
		}

		switch msg.Type {
		case 0:
			msg.Type = 1
			go this.Broadcast(msg)
			break
		case 1:
			err := storage.EDB.UpdateEndpoint(msg.WalletAddr, msg.HostPort)
			if err != nil {
				log.Errorf("err\n", err)
			}
			break
		default:
			log.Debugf("unknown msg type: %d\n", msg.Type)
		}
	case *pm.Torrent:
		log.Debug("dns actor receive pm.Torrent msg, InfoHash: %v, Left: %d, msg.PeerInfo: %v\n", msg.InfoHash, msg.Left, msg.Peerinfo)
		if msg.InfoHash == nil || msg.Peerinfo == nil {
			break
		}

		peer := storage.PeerInfo{}
		log.Debugf("PeerInfo Binary: %v\n", msg.Peerinfo)

		err := peer.Deserialize(msg.Peerinfo)
		peer.Print()
		if err != nil {
			log.Errorf("pm.Torrent Deserialize err: %v\n", err)
			break
		}

		switch msg.Type {
		case 0:
			msg.Type = 1
			go this.Broadcast(msg)
			break
		case 1:
			err := storage.TDB.AddTorrentPeer(msg.InfoHash, msg.Left, peer.NodeAddr.String(), &peer)
			if err != nil {
				log.Errorf("%v\n", err)
			}
			break
		default:
			log.Debugf("unknown msg type: %d\n", msg.Type)
		}
	default:
		log.Debugf("dns actor receive unknown message type.")
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
