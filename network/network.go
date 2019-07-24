package network

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/saveio/carrier/crypto"
	"github.com/saveio/carrier/crypto/ed25519"
	"github.com/saveio/carrier/network"
	"github.com/saveio/carrier/network/components/backoff"
	"github.com/saveio/carrier/network/components/keepalive"
	"github.com/saveio/carrier/network/components/proxy"
	"github.com/saveio/carrier/types/opcode"
	"github.com/saveio/dsp-go-sdk/network/common"
	act "github.com/saveio/pylons/actor/server"
	"github.com/saveio/themis/common/log"

	"github.com/saveio/pylons/common/constants"
	"github.com/saveio/pylons/network/transport/messages"
	"github.com/saveio/pylons/transfer"

	//"github.com/golang/protobuf/proto"
	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/scan/common/config"
	pm "github.com/saveio/scan/messages/protoMessages"
)

var DDNSP2P *Network

var once sync.Once

const (
	OpCodeProcessed opcode.Opcode = 1000 + iota
	OpCodeDelivered
	OpCodeSecretRequest
	OpCodeRevealSecret
	OpCodeSecretMsg
	OpCodeDirectTransfer
	OpCodeLockedTransfer
	OpCodeRefundTransfer
	OpCodeLockExpired
	OpCodeWithdrawRequest
	OpCodeWithdraw
	OpCodeCooperativeSettleRequest
	OpCodeCooperativeSettle
	OpCodeRegistry
	OpCodeUnRegistry
	OpCodeTorrent
)

var opCodes = map[opcode.Opcode]proto.Message{
	OpCodeRegistry:                 &pm.Registry{},
	OpCodeUnRegistry:               &pm.UnRegistry{},
	OpCodeTorrent:                  &pm.Torrent{},
	OpCodeProcessed:                &messages.Processed{},
	OpCodeDelivered:                &messages.Delivered{},
	OpCodeSecretRequest:            &messages.SecretRequest{},
	OpCodeRevealSecret:             &messages.RevealSecret{},
	OpCodeSecretMsg:                &messages.Secret{},
	OpCodeDirectTransfer:           &messages.DirectTransfer{},
	OpCodeLockedTransfer:           &messages.LockedTransfer{},
	OpCodeRefundTransfer:           &messages.RefundTransfer{},
	OpCodeLockExpired:              &messages.LockExpired{},
	OpCodeWithdrawRequest:          &messages.WithdrawRequest{},
	OpCodeWithdraw:                 &messages.Withdraw{},
	OpCodeCooperativeSettleRequest: &messages.CooperativeSettleRequest{},
	OpCodeCooperativeSettle:        &messages.CooperativeSettle{},
}

type Network struct {
	P2p                   *network.Network
	peerAddrs             []string
	listenAddr            string
	proxyAddr             string
	pid                   *actor.PID
	protocol              string
	address               string
	mappingAddress        string
	Keys                  *crypto.KeyPair
	keepaliveInterval     time.Duration
	keepaliveTimeout      time.Duration
	peerStateChan         chan *keepalive.PeerStateEvent
	kill                  chan struct{}
	ActivePeers           *sync.Map
	addressForHealthCheck *sync.Map
	Bootstraps            []string
}

func NewP2P() *Network {
	n := &Network{
		P2p: new(network.Network),
	}
	n.ActivePeers = new(sync.Map)
	n.addressForHealthCheck = new(sync.Map)
	n.kill = make(chan struct{})
	n.peerStateChan = make(chan *keepalive.PeerStateEvent, 16)
	return n
}

func (this *Network) SetProxyServer(address string) {
	this.proxyAddr = address
}

func (this *Network) Protocol() string {
	idx := strings.Index(this.PublicAddr(), "://")
	if idx == -1 {
		return "tcp"
	}
	return this.PublicAddr()[:idx]
}

func (this *Network) Start(address string, bootstraps []string) error {
	protocolIndex := strings.Index(address, "://")
	if protocolIndex == -1 {
		return errors.New("invalid address")
	}
	protocol := address[:protocolIndex]
	log.Debugf("channel protocol %s", protocol)
	builder := network.NewBuilderWithOptions(network.WriteFlushLatency(1 * time.Millisecond))
	if this.Keys != nil {
		log.Debugf("channel use account key")
		builder.SetKeys(this.Keys)
	} else {
		log.Debugf("channel use RandomKeyPair key")
		builder.SetKeys(ed25519.RandomKeyPair())
	}

	builder.SetAddress(address)
	if this.keepaliveInterval == 0 {
		this.keepaliveInterval = keepalive.DefaultKeepaliveInterval
	}
	if this.keepaliveTimeout == 0 {
		this.keepaliveTimeout = keepalive.DefaultKeepaliveTimeout
	}
	options := []keepalive.ComponentOption{
		keepalive.WithKeepaliveInterval(this.keepaliveInterval),
		keepalive.WithKeepaliveTimeout(this.keepaliveTimeout),
		keepalive.WithPeerStateChan(this.peerStateChan),
	}
	component := new(NetComponent)
	component.Net = this
	builder.AddComponent(component)
	builder.AddComponent(keepalive.New(options...))
	backoff_options := []backoff.ComponentOption{
		backoff.WithInitialDelay(3 * time.Second),
		backoff.WithMaxAttempts(10),
		backoff.WithPriority(10),
	}
	builder.AddComponent(backoff.New(backoff_options...))
	fmt.Println(protocol, this.proxyAddr)

	if len(this.proxyAddr) > 0 {
		switch protocol {
		case "udp":
			fmt.Println("UDPProxyComponent")
			builder.AddComponent(new(proxy.UDPProxyComponent))
		case "kcp":
			fmt.Println("KCPProxyComponent")
			builder.AddComponent(new(proxy.KCPProxyComponent))
		case "quic":
			fmt.Println("QuicProxyComponent")
			builder.AddComponent(new(proxy.QuicProxyComponent))
		case "tcp":
			fmt.Println("TcpProxyComponent")
			builder.AddComponent(new(proxy.TcpProxyComponent))
		}
	}
	var err error
	this.P2p, err = builder.Build()
	if err != nil {
		log.Error("[P2pNetwork] Start builder.Build error: ", err.Error())
		return err
	}

	once.Do(func() {
		for k, v := range opCodes {
			err := opcode.RegisterMessageType(k, v)
			if err != nil {
				panic("register messages failed")
			}
		}
	})
	if len(this.proxyAddr) > 0 {
		this.P2p.EnableProxyMode(true)
		this.P2p.SetProxyServer(this.proxyAddr)
	}
	this.P2p.SetNetworkID(config.DefaultConfig.CommonConfig.NetworkId)
	go this.P2p.Listen()
	go this.PeerStateChange(this.syncPeerState)

	this.P2p.BlockUntilListening()
	log.Debugf("channel will BlockUntilProxyFinish..., networkid %d", config.DefaultConfig.CommonConfig.NetworkId)
	if len(this.proxyAddr) > 0 {
		switch protocol {
		case "udp":
			this.P2p.BlockUntilUDPProxyFinish()
		case "kcp":
			this.P2p.BlockUntilKCPProxyFinish()
		case "quic":
			this.P2p.BlockUntilQuicProxyFinish()
		case "tcp":
			this.P2p.BlockUntilTcpProxyFinish()

		}
	}
	log.Debugf("finish BlockUntilProxyFinish...")
	if len(this.P2p.ID.Address) == 6 {
		return errors.New("invalid address")
	}
	return nil
}

func (this *Network) Halt() error {
	if this.P2p == nil {
		return errors.New("network is down")
	}
	this.P2p.Close()
	return nil
}

func (this *Network) Dial(addr string) error {
	if this.P2p == nil {
		return errors.New("network is nil")
	}
	_, err := this.P2p.Dial(addr)
	return err
}

func (this *Network) Disconnect(addr string) error {
	if this.P2p == nil {
		return errors.New("network is nil")
	}
	peer, err := this.P2p.Client(addr)
	if err != nil {
		return err
	}
	return peer.Close()
}

// IsPeerListenning. check the peer is listening or not.
func (this *Network) IsPeerListenning(addr string) bool {
	if this.P2p == nil {
		return false
	}
	err := this.Dial(addr)
	if err != nil {
		return false
	}
	err = this.Disconnect(addr)
	if err != nil {
		return false
	}
	return true
}

func (this *Network) Stop() {
	close(this.kill)
	this.P2p.Close()
}

func (this *Network) Connect(tAddr string) error {
	if _, ok := this.ActivePeers.Load(tAddr); ok {
		// node is active, no need to connect
		pse := &keepalive.PeerStateEvent{
			Address: tAddr,
			State:   keepalive.PEER_REACHABLE,
		}
		this.peerStateChan <- pse
		return nil
	}
	if _, ok := this.addressForHealthCheck.Load(tAddr); ok {
		// already try to connect, donn't retry before we get a result
		log.Infof("already try to connect %s", tAddr)
		return nil
	}
	this.addressForHealthCheck.Store(tAddr, struct{}{})
	this.P2p.Bootstrap(tAddr)
	return nil
}

func (this *Network) Close(tAddr string) error {
	peer, err := this.P2p.Client(tAddr)
	if err != nil {
		log.Error("[P2P Close] close addr: %s error ", tAddr)
	} else {
		this.addressForHealthCheck.Delete(tAddr)
		peer.Close()
	}
	return nil
}

// Send send msg to peer asyncnously
// peer can be addr(string) or client(*network.peerClient)
func (this *Network) Send(msg proto.Message, toAddr string) error {
	if _, ok := this.ActivePeers.Load(toAddr); !ok {
		return fmt.Errorf("can not send to inactive peer %s", toAddr)
	}
	signed, err := this.P2p.PrepareMessage(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to sign message")
	}
	log.Debugf("write msg to %s", toAddr)
	err = this.P2p.Write(toAddr, signed)
	log.Debugf("write msg done to %s", toAddr)
	if err != nil {
		return fmt.Errorf("failed to send message to %s", toAddr)
	}
	return nil
}
func (this *Network) ListenAddr() string {
	return this.listenAddr
}

func (this *Network) PublicAddr() string {
	return this.P2p.ID.Address
}

func (this *Network) GetPeersIfExist() error {
	this.P2p.EachPeer(func(client *network.PeerClient) bool {
		this.peerAddrs = append(this.peerAddrs, client.Address)
		return true
	})
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

func (this *Network) Request(msg proto.Message, peer string) (proto.Message, error) {
	client := this.P2p.GetPeerClient(peer)
	if client == nil {
		return nil, fmt.Errorf("get peer client is nil %s", peer)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(common.REQUEST_MSG_TIMEOUT)*time.Second)
	defer cancel()
	return client.Request(ctx, msg)
}

func (this *Network) RequestWithRetry(msg proto.Message, peer string, retry int) (proto.Message, error) {
	client := this.P2p.GetPeerClient(peer)
	if client == nil {
		return nil, fmt.Errorf("get peer client is nil %s", peer)
	}
	var res proto.Message
	var err error
	for i := 0; i < retry; i++ {
		log.Debugf("send request msg to %s with retry %d", peer, retry)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(common.REQUEST_MSG_TIMEOUT)*time.Second)
		defer cancel()
		res, err = client.Request(ctx, msg)
		if err == nil || err.Error() != "context deadline exceeded" {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *Network) PeerStateChange(fn func(*keepalive.PeerStateEvent)) {
	ka, reg := this.P2p.Component(keepalive.ComponentID)
	if !reg {
		log.Error("keepalive component do not reg")
		return
	}
	peerStateChan := ka.(*keepalive.Component).GetPeerStateChan()
	stopCh := ka.(*keepalive.Component).GetStopChan()
	for {
		select {
		case event := <-peerStateChan:
			fn(event)

		case <-stopCh:
			return

		}
	}
}
func (this *Network) syncPeerState(state *keepalive.PeerStateEvent) {
	var nodeNetworkState string
	log.Debugf("[syncPeerState] addr: %s state: %v", state.Address, state.State)
	switch state.State {
	case keepalive.PEER_REACHABLE:
		log.Debugf("[syncPeerState] addr: %s state: NetworkReachable\n", state.Address)
		this.ActivePeers.LoadOrStore(state.Address, struct{}{})
		nodeNetworkState = transfer.NetworkReachable
		act.SetNodeNetworkState(state.Address, nodeNetworkState)
	case keepalive.PEER_UNKNOWN:
		log.Debugf("[syncPeerState] addr: %s state: PEER_UNKNOWN\n", state.Address)
	case keepalive.PEER_UNREACHABLE:
		this.ActivePeers.Delete(state.Address)
		log.Debugf("[syncPeerState] addr: %s state: NetworkUnreachable\n", state.Address)
		nodeNetworkState = transfer.NetworkUnreachable
		act.SetNodeNetworkState(state.Address, nodeNetworkState)
	}
}

//P2P network msg receive. torrent msg, reg msg, unReg msg
func (this *Network) Receive(message proto.Message, from string) error {
	switch message.(type) {
	case *messages.Processed:
		act.OnBusinessMessage(message, from)
	case *messages.Delivered:
		act.OnBusinessMessage(message, from)
	case *messages.SecretRequest:
		act.OnBusinessMessage(message, from)
	case *messages.RevealSecret:
		act.OnBusinessMessage(message, from)
	case *messages.Secret:
		act.OnBusinessMessage(message, from)
	case *messages.DirectTransfer:
		act.OnBusinessMessage(message, from)
	case *messages.LockedTransfer:
		act.OnBusinessMessage(message, from)
	case *messages.RefundTransfer:
		act.OnBusinessMessage(message, from)
	case *messages.LockExpired:
		act.OnBusinessMessage(message, from)
	case *messages.WithdrawRequest:
		act.OnBusinessMessage(message, from)
	case *messages.Withdraw:
		act.OnBusinessMessage(message, from)
	case *messages.CooperativeSettleRequest:
		act.OnBusinessMessage(message, from)
	case *messages.CooperativeSettle:
		act.OnBusinessMessage(message, from)
	case *pm.Torrent:
		log.Errorf("[MSB Receive] receive from peer:%s, nil Torrent message", from)
		this.OnBusinessMessage(message, from)
	case *pm.Registry:
		log.Errorf("[MSB Receive] receive from peer:%s, nil Reg message", from)
		this.OnBusinessMessage(message, from)
	case *pm.UnRegistry:
		log.Errorf("[MSB Receive] receive from peer:%s, nil Unreg message", from)
		this.OnBusinessMessage(message, from)

	default:
		// log.Errorf("[MSB Receive] unknown message type:%s", msg.String())
	}
	return nil
}

func (this *Network) OnBusinessMessage(message proto.Message, from string) error {
	log.Debugf("[OnBusinessMessage] receive message from peer:%s", from)
	future := this.GetPID().RequestFuture(message,
		constants.REQ_TIMEOUT*time.Second)
	if _, err := future.Result(); err != nil {
		log.Error("[OnBusinessMessage] error: ", err)
		return err
	}
	return nil
}
