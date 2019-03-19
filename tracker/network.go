/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-03-12 
*/
package tracker

import (
	"github.com/oniio/oniP2p/network"
	"github.com/oniio/oniP2p/crypto/ed25519"

	"github.com/oniio/oniChain/common/log"
	"github.com/oniio/oniP2p/network/keepalive"
	"github.com/oniio/oniP2p/network/nat"
	"github.com/oniio/oniDNS/config"
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/oniio/oniP2p/types/opcode"
	m "github.com/oniio/oniDNS/tracker/messages"
	"github.com/oniio/oniDNS/tracker/common"
	"fmt"
)

type Network struct {
	*network.Component
	n          *network.Network
	peerAddrs  []string
	listenAddr string
	*Server
}

func (this *Network) Start() error {
	keys := ed25519.RandomKeyPair()
	builder := network.NewBuilder()
	builder.SetKeys(keys)
	builder.SetAddress(network.FormatAddress("udp", "127.0.0.1", uint16(config.DefaultConfig.Tracker.SyncPort)))
	opcode.RegisterMessageType(opcode.Opcode(common.SYNC_MSG_OP_CODE), &m.SyncMessage{})
	opcode.RegisterMessageType(opcode.Opcode(common.SYNC_REGMSG_OP_CODE), &m.Registry{})
	opcode.RegisterMessageType(opcode.Opcode(common.SYNC_UNREGMSG_OP_CODE), &m.UnRegistry{})
	peerStateChan := make(chan *keepalive.PeerStateEvent, 10)
	options := []keepalive.ComponentOption{
		keepalive.WithKeepaliveInterval(keepalive.DefaultKeepaliveInterval),
		keepalive.WithKeepaliveTimeout(keepalive.DefaultKeepaliveTimeout),
		keepalive.WithPeerStateChan(peerStateChan),
	}
	builder.AddComponent(keepalive.New(options...))
	builder.AddComponentWithPriority(-9998, new(nat.StunComponent))
	net, err := builder.Build()
	this.n = net
	if err != nil {
		log.Fatal(err)
		return err
	}
	go this.n.Listen()
	this.n.BlockUntilListening()
	sc, reg := this.n.Component(nat.StunComponentID)
	if !reg {
		log.Error("stun component don't reg ")
	} else {
		exAddr := sc.(*nat.StunComponent).GetPublicAddr()
		this.n.ExternalAddr = exAddr
	}
	log.Infof("Listening for peers on %s.", this.n.ExternalAddr)
	peers := config.DefaultConfig.Tracker.SeedLists
	if len(peers) > 0 {
		this.n.Bootstrap(peers...)
	}
	return nil
}

func (this *Network) NewNetwork() *Network {
	return new(Network)

}

func (this *Network) ListenAddr() string {
	return this.listenAddr
}

func (this *Network) Receive(ctx *network.ComponentContext) error {
	switch msg := ctx.Message().(type) {
	case *m.SyncMessage:
		k := msg.InfoHash
		v := msg.Torrent

		if err:=this.ls.Put(k, v);err!=nil{
			return fmt.Errorf("[Receive] sync filemessage error:%v",err)
		}

	case *m.Registry:
		walletAddr := msg.WalletAddr
		hostPort:=msg.HostPort

		if err:=RegMsgDB.Put(walletAddr,hostPort);err!=nil{
			return fmt.Errorf("[Receive] sync regmessage error:%v",err)
		}

	case *m.UnRegistry:
		walletAddr:=msg.WalletAddr
		if err:=RegMsgDB.Delete(walletAddr);err!=nil{
			return fmt.Errorf("[Receive] sync unregmessage error:%v",err)
		}
	default:
		return fmt.Errorf("[Receive] unknown message type:%v",msg)

	}
	return nil
}

func (this *Network) GetPeersIfExist() error {
	this.n.EachPeer(func(client *network.PeerClient) bool {
		this.peerAddrs = append(this.peerAddrs, client.Address)
		return true
	})
	return nil
}

func (this *Network) Connect(addr ...string) error {

	return nil
}

func (this *Network) BroadCast(ctx context.Context, message proto.Message) error {
	this.n.Broadcast(ctx, message)
	return nil
}

func (this *Network) BroadCastByaddr(addr ...string) error {

	return nil
}

func (this *Network) BroadCastRandom(addr ...string) error {

	return nil
}

func (this *Network) SyncTorrent(k, v []byte) error {
	ms := &m.SyncMessage{}
	ms.InfoHash = k
	ms.Torrent = v
	ctx := network.WithSignMessage(context.Background(), true)

	return this.BroadCast(ctx, ms)
}

func (this *Network) SyncRegMsg(walletAddr, hostPort string) error {
	rm := &m.Registry{}
	rm.WalletAddr = walletAddr
	rm.HostPort = hostPort

	ctx := network.WithSignMessage(context.Background(), true)
	return this.BroadCast(ctx, rm)
}

func (this *Network) SyncUnRegMsg(walletAddr string) error {
	rm := &m.UnRegistry{}
	rm.WalletAddr = walletAddr
	ctx := network.WithSignMessage(context.Background(), true)
	return this.BroadCast(ctx, rm)
}