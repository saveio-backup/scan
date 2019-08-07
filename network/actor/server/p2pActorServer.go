package server

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/saveio/scan/channel"
	"github.com/saveio/scan/storage"

	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	p2pNet "github.com/saveio/carrier/network"
	dspact "github.com/saveio/dsp-go-sdk/actor/client"
	chact "github.com/saveio/pylons/actor/client"
	comm "github.com/saveio/scan/common"
	"github.com/saveio/scan/common/config"
	pm "github.com/saveio/scan/messages/protoMessages"
	"github.com/saveio/scan/network"
	"github.com/saveio/scan/network/actor/messages"
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
	localPid, err := actor.SpawnNamed(this.props, "scan_net_server")
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
	case *pm.Registry:
		log.Debugf("tracker client registry or update msg:%s", msg.String())
		if msg.WalletAddr == "" || msg.HostPort == "" {
			log.Errorf("[MSB Receive] receive from peer:%s, nil Reg message", ctx.Sender().Address)
			break
		}
		k, v := comm.WHPTobyte(msg.WalletAddr, msg.HostPort)

		hpBytes, err := storage.TDB.Get(k)
		if msg.Type == 0 {
			msg.Type = 1
			go this.Broadcast(msg)
		} else if msg.Type == 1 && string(hpBytes) != string(v) {
			err = storage.TDB.Put(k, v)
			if err != nil {
				log.Errorf("[MSB Receive] sync registry failed, error:%v", err)
			}
			log.Infof("sethostaddr, walletAddr:%s, hostport:%s", msg.WalletAddr, msg.HostPort)
			channel.GlbChannelSvr.Channel.SetHostAddr(msg.WalletAddr, fmt.Sprintf("%s://%s", config.Parameters.Base.ChannelProtocol, msg.HostPort))

			log.Infof("[MSB Receive] sync registry success: %v", msg.String())
		} else {
			log.Debugf("No NEED BROADCAST, hpBytes: %s, v: %s", hpBytes, v)
		}

	case *pm.UnRegistry:
		log.Debugf("tracker client unRegistry msg:%s", msg.String())
		if msg.WalletAddr == "" {
			log.Errorf("[MSB Receive] receive from peer:%s, nil Reg message", ctx.Sender().Address)
			break
		}
		k := comm.WTobyte(msg.WalletAddr)

		hpBytes, err := storage.TDB.Get(k)
		if msg.Type == 0 {
			msg.Type = 1
			go this.Broadcast(msg)
		} else if msg.Type == 1 && string(hpBytes) != "" {
			err = storage.TDB.Delete(k)
			if err != nil {
				log.Errorf("[MSB Receive] sync regmessage error:%v", err)
			}
		} else {
			log.Errorf("No NEED BROADCAST, hpBytes: %s", hpBytes)
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
