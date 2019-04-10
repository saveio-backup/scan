/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-03-12 
*/
package network

import (
	"github.com/oniio/oniP2p/network"
	"github.com/oniio/oniP2p/crypto/ed25519"

	"github.com/oniio/oniChain/common/log"
	"github.com/oniio/oniP2p/network/keepalive"
	"github.com/oniio/oniP2p/network/nat"
	"github.com/oniio/oniDNS/config"
	"context"
	//"github.com/golang/protobuf/proto"
	"github.com/gogo/protobuf/proto"
	"github.com/oniio/oniP2p/types/opcode"
	pm "github.com/oniio/oniDNS/messages/protoMessages"
	"github.com/oniio/oniDNS/tracker/common"
	comm "github.com/oniio/oniDNS/common"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/oniio/oniDNS/storage"
	"fmt"
	"time"
	"errors"
)

var DDNSP2P *Network

type Network struct {
	*network.Component
	*network.Network
	peerAddrs  []string
	listenAddr string
	pid *actor.PID
}

func NewP2P()*Network{
	n:=&Network{
		Network:new(network.Network),
	}
	return n

}

func (this *Network) Start() error {
	keys := ed25519.RandomKeyPair()
	builder := network.NewBuilder()
	builder.SetKeys(keys)
	builder.SetAddress(network.FormatAddress("udp", "127.0.0.1", uint16(config.DefaultConfig.Tracker.SyncPort)))
	opcode.RegisterMessageType(opcode.Opcode(common.SYNC_MSG_OP_CODE), &pm.Torrent{})
	opcode.RegisterMessageType(opcode.Opcode(common.SYNC_REGMSG_OP_CODE), &pm.Registry{})
	opcode.RegisterMessageType(opcode.Opcode(common.SYNC_UNREGMSG_OP_CODE), &pm.UnRegistry{})
	peerStateChan := make(chan *keepalive.PeerStateEvent, 10)
	options := []keepalive.ComponentOption{
		keepalive.WithKeepaliveInterval(keepalive.DefaultKeepaliveInterval),
		keepalive.WithKeepaliveTimeout(keepalive.DefaultKeepaliveTimeout),
		keepalive.WithPeerStateChan(peerStateChan),
	}
	builder.AddComponent(keepalive.New(options...))
	builder.AddComponentWithPriority(-9998, new(nat.StunComponent))
	net, err := builder.Build()
	this.Network = net
	common.ListeningCh=make(chan struct{})
	close(common.ListeningCh)
	if err != nil {
		log.Fatal(err)
		return err
	}
	go this.Listen()
	this.BlockUntilListening()
	sc, reg := this.Network.Component(nat.StunComponentID)
	if !reg {
		log.Error("stun component don't reg ")
	} else {
		exAddr := sc.(*nat.StunComponent).GetPublicAddr()
		this.ExternalAddr = exAddr
	}
	log.Infof("Listening for peers on %s.", this.ExternalAddr)
	peers := config.DefaultConfig.Tracker.SeedLists
	if len(peers) > 0 {
		this.Bootstrap(peers...)
		log.Debug("had bootStraped peers")
	}
	comm.WaitToExit()
	return nil
}

//P2P network msg receive. torrent msg, reg msg, unReg msg
func (this *Network) Receive(ctx *network.ComponentContext) error {
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
			k, v := comm.WHPTobyte(msg.WalletAddr, msg.HostPort)

			if err := storage.TDB.Put(k, v); err != nil {
				log.Errorf("[MSB Receive] sync regmessage error:%v", err)
			}

		case *pm.UnRegistry:
			k, _ := comm.WHPTobyte(msg.WalletAddr, "")
			if err := storage.TDB.Delete(k); err != nil {
				return fmt.Errorf("[MSB Receive] sync unregmessage error:%v", err)
			}
		default:
			log.Errorf("[MSB Receive] unknown message type:%s", msg.String())

		}
	}
	return nil
}

func (this *Network) NewNetwork() *Network {
	return new(Network)

}

func (this *Network) ListenAddr() string {
	return this.listenAddr
}

func (this *Network) GetPeersIfExist() error {
	this.EachPeer(func(client *network.PeerClient) bool {
		this.peerAddrs = append(this.peerAddrs, client.Address)
		return true
	})
	return nil
}

func (this *Network) Connect(tAddr ...string) error {
	this.Network.Bootstrap(tAddr...)
	for _, addr := range tAddr {
		exist := this.Network.ConnectionStateExists(addr)
		if !exist {
			return fmt.Errorf("[P2P connect] bootstrap addr:%s error",addr)
		}
	}
	return nil
}

// Send send msg to peer asyncnously
// peer can be addr(string) or client(*network.peerClient)
func (this *Network) Send(msg proto.Message, peer interface{}) error {
	client, err := this.loadClient(peer)
	if err != nil {
		return err
	}
	return client.Tell(context.Background(), msg)
}

// Request. send msg to peer and wait for response synchronously
func (this *Network) Request(msg proto.Message, peer interface{},timeout uint64) (proto.Message, error) {
	client, err := this.loadClient(peer)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	res, err := client.Request(ctx, msg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *Network) BroadCast(ctx context.Context, message proto.Message) error {
	this.Broadcast(ctx, message)
	return nil
}

// SetPID sets p2p actor
func (this *Network) SetPID(pid *actor.PID) {
	this.pid = pid
	//this.msgRouter.SetPID(pid)
}

// GetPID returns p2p actor
func (this *Network) GetPID() *actor.PID {
	return this.pid
}

func (this *Network) loadClient(peer interface{}) (*network.PeerClient, error) {
	addr, ok := peer.(string)
	if ok {
		client, err := this.Network.Client(addr)
		if err != nil {
			return nil, err
		}
		if client == nil {
			return nil, errors.New("client is nil")
		}
		return client, nil
	}
	client, ok := peer.(*network.PeerClient)
	if !ok || client == nil {
		return nil, errors.New("invalid peer type")
	}
	return client, nil
}

func (this *Network)ConnState(address string)(*network.ConnState,bool){
	return this.ConnectionState(address)
}

func (this *Network)ConnStateExists(address string)(*network.ConnState,bool){
	return this.ConnStateExists(address)
}

