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

type msgBus struct {
	*network.SyncNetwork
	MsgBox   chan *RegMsg
	TNTBox   chan *TorrenMsg
	UnMsgBox chan *UnRegMsg
}

func NewMsgBus() *msgBus {
	return &msgBus{
		SyncNetwork:new(network.SyncNetwork),
		MsgBox: make(chan *RegMsg),
		TNTBox: make(chan *TorrenMsg),
		UnMsgBox: make(chan *UnRegMsg),

	}
}

//msgBus start and local msg receive.
func (this *msgBus) Start() error {
	go this.SyncNetwork.Start()
	go this.start()
	return nil
}

func (this *msgBus) start() error {
	log.Info("msgBus is accepting for local channels ")
	for {
		select {
		case m := <-this.TNTBox:
			this.localTntHandler(m)

		case m := <-this.MsgBox:
			this.localMsgHandler(m)

		case m:=<- this.UnMsgBox:
			this.localUnRegHandler(m)
		}
	}
}

//syncNet msg receive. torrent msg, reg msg, unReg msg
func (this *msgBus) Receive(ctx *p2pNet.ComponentContext) error {
	return this.syncMsgHandler(ctx)
}

//the relative handle with msg from syncNet
func (this *msgBus) syncMsgHandler(ctx *p2pNet.ComponentContext) error {
	log.Info("msgBus is accepting for syncNet messages ")
	for {
		switch msg := ctx.Message().(type) {
		case *pm.Torrent:
			k := msg.InfoHash
			v := msg.Torrent

			if err := storage.TDB.Put(k, v); err != nil {
				return fmt.Errorf("[Receive] sync filemessage error:%v", err)
			}

		case *pm.Registry:
			k, v := common.WHPTobyte(msg.WalletAddr, msg.HostPort)

			if err := storage.TDB.Put(k, v); err != nil {
				return fmt.Errorf("[Receive] sync regmessage error:%v", err)
			}

		case *pm.UnRegistry:
			k, _ := common.WHPTobyte(msg.WalletAddr, "")
			if err := storage.TDB.Delete(k); err != nil {
				return fmt.Errorf("[Receive] sync unregmessage error:%v", err)
			}
		default:
			return fmt.Errorf("[MSB Receive] unknown message type:%s", msg.String())

		}
	}
	return nil
}

//the relative handle with msg from local api. (local reg and unReg msg, sync to peer)
func (this *msgBus) localMsgHandler(msg *RegMsg) error {
	if msg == nil {
		return errors.NewErr("[localTntHandler] msg is nil")
	}
	m:=&pm.Registry{
		WalletAddr:msg.WalletAddr,
		HostPort:msg.HostPort,
	}
	ctx := p2pNet.WithSignMessage(context.Background(), true)
	this.Network.Broadcast(ctx, m)

	return nil
}

func (this *msgBus) localUnRegHandler(msg *UnRegMsg) error {
	if msg == nil {
		return errors.NewErr("[localTntHandler] msg is nil")
	}
	m:=&pm.UnRegistry{
		WalletAddr:msg.WalletAddr,
	}
	ctx := p2pNet.WithSignMessage(context.Background(), true)
	this.Network.Broadcast(ctx, m)

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
	return nil

}
