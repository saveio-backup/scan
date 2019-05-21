/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-03-31
 */
package recv

import (
	"context"
	"fmt"
	"reflect"

	"github.com/saveio/scan/storage"

	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	p2pNet "github.com/saveio/carrier/network"
	act "github.com/saveio/pylons/actor/client"
	comm "github.com/saveio/scan/common"
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

func NewP2PActor(n *network.Network) (*actor.PID, error) {
	var err error
	p2pActor := &P2PActor{
		net:         n,
		msgHandlers: make(map[string]MessageHandler),
	}
	p2pActor.localPID, err = p2pActor.Start()
	if err != nil {
		return nil, err
	}
	return p2pActor.localPID, nil

}

func (this *P2PActor) Start() (*actor.PID, error) {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	localPid, err := actor.SpawnNamed(this.props, "net_server")
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
	case *act.ConnectReq:
		log.Warn("[P2p]actor ConnectReq Address: %s", msg.Address)
		err := this.net.Connect(msg.Address)
		fmt.Println(err)
		ctx.Sender().Request(&act.P2pResp{err}, ctx.Self())
	case *act.CloseReq:
		log.Warn("[P2p]actor CloseReq")
		err := this.net.Close(msg.Address)
		ctx.Sender().Request(&act.P2pResp{err}, ctx.Self())
	case *act.SendReq:
		log.Warn("[P2p]actor SendReq")
		err := this.net.Send(msg.Data, msg.Address)
		ctx.Sender().Request(&act.P2pResp{err}, ctx.Self())
	case *messages.Ping:
		log.Warn("[P2p]actor Ping")
		ctx.Sender().Tell(&messages.Pong{})
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
				log.Errorf("[MSB Receive] sync regmessage error:%v", err)
			}
		} else {
			log.Errorf("No NEED BROADCAST, hpBytes: %s, v: %s", hpBytes, v)
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
