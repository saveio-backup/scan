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
	"github.com/golang/protobuf/proto"
	"github.com/oniio/oniP2p/types/opcode"
	pm "github.com/oniio/oniDNS/messageBus/protoMessages"
	"github.com/oniio/oniDNS/tracker/common"
)

type SyncNetwork struct {
	*network.Component
	*network.Network
	peerAddrs  []string
	listenAddr string
}

func (this *SyncNetwork) Start() error {
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

	return nil
}

func (this *SyncNetwork) NewNetwork() *SyncNetwork {
	return new(SyncNetwork)

}

func (this *SyncNetwork) ListenAddr() string {
	return this.listenAddr
}


func (this *SyncNetwork) GetPeersIfExist() error {
	this.EachPeer(func(client *network.PeerClient) bool {
		this.peerAddrs = append(this.peerAddrs, client.Address)
		return true
	})
	return nil
}

func (this *SyncNetwork) Connect(addr ...string) error {

	return nil
}

func (this *SyncNetwork) BroadCast(ctx context.Context, message proto.Message) error {
	this.Broadcast(ctx, message)
	return nil
}
