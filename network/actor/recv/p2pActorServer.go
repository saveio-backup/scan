/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-03-31 
*/
package recv

import (
	"github.com/oniio/oniDNS/network"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/oniio/oniChain/common/log"
	"github.com/ontio/ontology-eventbus/remote"
	"github.com/gogo/protobuf/proto"
	"reflect"
	"github.com/oniio/oniDNS/network/actor/messages"
	p2pNet "github.com/oniio/oniP2p/network"
	"context"
	pm "github.com/oniio/oniDNS/messages/protoMessages"
	"fmt"
)
type MessageHandler func(msgData interface{},pid *actor.PID)
var P2pPid *actor.PID
type P2PActor struct {
	net *network.Network
	props *actor.Props
	msgHandlers map[string]MessageHandler
	localPID *actor.PID
}

func NewP2PActor(n *network.Network)error{
	var err error
	p2pActor:= &P2PActor{
		net:n,
		msgHandlers:make(map[string]MessageHandler),
	}
	p2pActor.localPID,err=p2pActor.Start()
	if err!=nil{
		return err
	}
	return nil

}

func (this *P2PActor) Start() (*actor.PID, error) {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	localPid, err := actor.SpawnNamed(this.props, "net_server")
	if err!=nil{
		return nil,fmt.Errorf("[P2PActor] start error:%v",err)
	}
	this.localPID=localPid
	return localPid, err
}

func (this *P2PActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Restarting:
		log.Warn("[oniP2p]actor restarting")
	case *actor.Stopping:
		log.Warn("[oniP2p]actor stopping")
	case *actor.Stopped:
		log.Warn("[oniP2p]actor stopped")
	case *actor.Started:
		log.Debug("[oniP2p]actor started")
	case *actor.Restart:
		log.Warn("[oniP2p]actor restart")
	case *messages.Ping:
		ctx.Sender().Tell(&messages.Pong{})
	case *pm.Registry:
		go this.Broadcast(msg)
	case *pm.UnRegistry:
		go this.Broadcast(msg)
	case *pm.Torrent:
		go this.Broadcast(msg)
	case *messages.UserDefineMsg:
		t:=reflect.TypeOf(msg)
		this.msgHandlers[t.Name()](msg,this.localPID)
		
	default:
		log.Error("[P2PActor] receive unknown message type!")
	}

}

func (this *P2PActor)Broadcast(message proto.Message){
	ctx := p2pNet.WithSignMessage(context.Background(), true)
	this.net.Broadcast(ctx,message)

}

func (this *P2PActor)RegMsgHandler(msgName string,handler MessageHandler){
	this.msgHandlers[msgName]=handler
}

func (this *P2PActor)UnRegMsgHandler(msgName string,handler MessageHandler){
	delete(this.msgHandlers,msgName)
}

func (this *P2PActor)SetLocalPID(pid *actor.PID){
	this.localPID=pid
}

func (this *P2PActor)GetLocalPID() *actor.PID{
	return this.localPID
}

//start local remote actor grpc service
func RemoteStart(addr string){
	remote.Start(addr)
}

//return remote actor grpc service pid
func RemotePid(addr,desc string)*actor.PID{
	if desc==""{
		desc="remote"
	}
	return actor.NewPID(addr,desc)
}