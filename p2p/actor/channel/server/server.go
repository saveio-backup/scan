package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	p2pNet "github.com/saveio/carrier/network"
	dspact "github.com/saveio/dsp-go-sdk/actor/client"
	chact "github.com/saveio/pylons/actor/client"
	"github.com/saveio/scan/p2p/actor/messages"
	network "github.com/saveio/scan/p2p/networks/channel"
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
			msg.Ret.Err = this.net.Connect(msg.Address)
			msg.Ret.Done <- true
		}()
	case *chact.CloseReq:
		go func() {
			msg.Ret.Err = this.net.Close(msg.Address)
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
	case *dspact.PeerListeningReq:
		ret := this.net.IsPeerListenning(msg.Address)
		var err error
		if !ret {
			err = errors.New("peer is not listening")
		}
		log.Debugf("is peer listening %s, ret %t, err %s", msg.Address, ret, err)
		ctx.Sender().Request(&dspact.P2pResp{Error: err}, ctx.Self())
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
