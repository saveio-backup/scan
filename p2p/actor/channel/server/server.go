package server

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	p2pNet "github.com/saveio/carrier/network"
	chact "github.com/saveio/pylons/actor/client"
	"github.com/saveio/scan/p2p/actor/messages"
	"github.com/saveio/scan/p2p/network"
	"github.com/saveio/themis/common/log"
)

var P2pPid *actor.PID

type MessageHandler func(msgData interface{}, pid *actor.PID)

type P2PActor struct {
	net                   *network.Network
	props                 *actor.Props
	msgHandlers           map[string]MessageHandler
	localPID              *actor.PID
	dnsHostAddrFromWallet func(string) string
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

func (this *P2PActor) SetDNSHostAddrFromWallet(f func(string) string) {
	this.dnsHostAddrFromWallet = f
}

func (this *P2PActor) Start() (*actor.PID, error) {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	localPid, err := actor.SpawnNamed(this.props, "scan_channel_net_server")
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
	case *chact.ConnectReq:
		go func() {
			if this.dnsHostAddrFromWallet == nil {
				msg.Ret.Done <- false
				return
			}
			hostAddr := this.dnsHostAddrFromWallet(msg.Address)
			log.Debugf("connect %s", hostAddr)
			_, msg.Ret.Err = this.net.Connect(hostAddr)
			msg.Ret.Done <- true
		}()
	case *chact.SendReq:
		go func() {
			msg.Ret.Err = this.net.Send(msg.Data, msg.Address)
			msg.Ret.Done <- true
		}()
	case *chact.GetNodeNetworkStateReq:
		go func() {
			state, err := this.net.GetPeerStateByAddress(msg.Address)
			msg.Ret.State = int(state)
			msg.Ret.Err = err
			msg.Ret.Done <- true
		}()
	case *messages.Ping:
		log.Warn("[P2p]actor Ping")
		ctx.Sender().Tell(&messages.Pong{})
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
