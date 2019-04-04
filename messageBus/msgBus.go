/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-03-24 
*/
package messageBus

import (
	"github.com/oniio/oniDNS/network"
	p2pNet "github.com/oniio/oniP2p/network"
	"context"
	"fmt"
	pm "github.com/oniio/oniDNS/messageBus/protoMessages"
	"github.com/oniio/oniDNS/storage"
	"github.com/oniio/oniDNS/common"
	"github.com/oniio/oniChain/errors"
	"github.com/oniio/oniChain/common/log"

	"github.com/ontio/ontology-eventbus/actor"
)

var MsgBus *msgBus

type TorrenMsg struct {
	InfoHash     []byte
	BytesTorrent []byte
}

type RegMsg struct {
	WalletAddr string
	HostPort   string
}

type UnRegMsg struct {
	WalletAddr string
}

type msgBusActor struct{}

type msgBus struct {
	*network.Network
	MsgBox   chan *RegMsg
	TNTBox   chan *TorrenMsg
	UnMsgBox chan *UnRegMsg
	p2p      *actor.PID // P2P actor
	localPid *actor.PID
}

func NewMsgBus() *msgBus {
	msgB := &msgBus{
		Network:  new(network.Network),
		MsgBox:   make(chan *RegMsg),
		TNTBox:   make(chan *TorrenMsg),
		UnMsgBox: make(chan *UnRegMsg),
	}
	props := actor.FromProducer(func() actor.Actor {
		return &msgBusActor{}
	})
	pid, err := actor.SpawnNamed(props, "msgBus")
	if err != nil {
		log.Errorf("[NewMsgBus] pid create error:%v", err)
		return nil
	}
	msgB.localPid = pid
	return msgB
}

//msgBus start and local msg receive.
func (this *msgBus) Start() error {
	go this.Network.Start()
	go this.accept()
	return nil
}

func (this *msgBus) accept() error {
	log.Info("msgBus is accepting for local channels ")
	for {
		select {
		case m := <-this.TNTBox:
			this.localTntHandler(m)

		case m := <-this.MsgBox:
			this.localMsgHandler(m)

		case m := <-this.UnMsgBox:
			this.localUnRegHandler(m)
		}
	}
}

//P2P network msg receive. torrent msg, reg msg, unReg msg
func (this *msgBus) Receive(ctx *p2pNet.ComponentContext) error {
	return this.syncMsgHandler(ctx)
}

//actor msg receive
func (this *msgBusActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Restarting:
		log.Warn("msgRouter actor restarting", msg)
	case *actor.Stopping:
		log.Warn("msgRouter actor stopping")
	case *actor.Stopped:
		log.Warn("msgRouter actor stopped")
	case *actor.Started:
		log.Warn("msgRouter actor started")
	case *actor.Restart:
		log.Warn("msgRouter actor restart")
	}
}

//the relative handle with msg from syncNet
func (this *msgBus) syncMsgHandler(ctx *p2pNet.ComponentContext) error {
	log.Info("msgBus is accepting for syncNet messages ")
	for {
		switch msg := ctx.Message().(type) {
		case *pm.Torrent:
			if msg.InfoHash == nil || msg.Torrent == nil {
				log.Errorf("[MSB Receive] receive from peer:%s, nil Torrent message",ctx.Sender().Address)
				break
			}
			k := msg.InfoHash
			v := msg.Torrent

			if err := storage.TDB.Put(k, v); err != nil {
				log.Errorf("[MSB Receive] sync filemessage error:%v", err)
			}

		case *pm.Registry:
			if msg.WalletAddr == "" || msg.HostPort == "" {
				log.Errorf("[MSB Receive] receive from peer:%s, nil Reg message", ctx.Sender().Address)
				break
			}
			k, v := common.WHPTobyte(msg.WalletAddr, msg.HostPort)

			if err := storage.TDB.Put(k, v); err != nil {
				log.Errorf("[MSB Receive] sync regmessage error:%v", err)
			}

		case *pm.UnRegistry:
			k, _ := common.WHPTobyte(msg.WalletAddr, "")
			if err := storage.TDB.Delete(k); err != nil {
				return fmt.Errorf("[MSB Receive] sync unregmessage error:%v", err)
			}
		default:
			log.Errorf("[MSB Receive] unknown message type:%s", msg.String())

		}
	}
	return nil
}

//the relative handle with msg from local api. (local reg and unReg msg, sync to peer)
func (this *msgBus) localMsgHandler(msg *RegMsg) error {
	if msg == nil {
		return errors.NewErr("[localTntHandler] msg is nil")
	}
	m := &pm.Registry{
		WalletAddr: msg.WalletAddr,
		HostPort:   msg.HostPort,
	}
	//ctx := p2pNet.WithSignMessage(context.Background(), true)
	//this.Network.Broadcast(ctx, m)
	this.p2p.Tell(m)
	return nil
}

func (this *msgBus) localUnRegHandler(msg *UnRegMsg) error {
	if msg == nil {
		return errors.NewErr("[localTntHandler] msg is nil")
	}
	m := &pm.UnRegistry{
		WalletAddr: msg.WalletAddr,
	}
	//ctx := p2pNet.WithSignMessage(context.Background(), true)
	//this.Network.Broadcast(ctx, m)
	this.p2p.Tell(m)

	return nil
}

//the relative handle with msg from local tracker server.(local torrent, sync to peer)
func (this *msgBus) localTntHandler(msg *TorrenMsg) error {
	if msg == nil {
		return errors.NewErr("[localTntHandler] msg is nil")
	}
	m := &pm.Torrent{
		InfoHash: msg.InfoHash,
		Torrent:  msg.BytesTorrent,
	}
	ctx := p2pNet.WithSignMessage(context.Background(), true)
	this.Network.Broadcast(ctx, m)
	this.p2p.Tell(m)
	return nil
}

// SetPID sets p2p actor
func (this *msgBus) SetPID(pid *actor.PID) {
	this.p2p = pid
}

// GetPID return msgBus actor
func (this *msgBus) GetPID() *actor.PID {
	return this.p2p
}
