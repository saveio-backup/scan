package dns

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/carrier/crypto"
	"github.com/saveio/carrier/crypto/ed25519"
	"github.com/saveio/carrier/network"
	"github.com/saveio/carrier/network/components/keepalive"
	"github.com/saveio/carrier/network/components/proxy"
	"github.com/saveio/carrier/types/opcode"
	"github.com/saveio/dsp-go-sdk/network/common"
	"github.com/saveio/scan/common/config"
	pm "github.com/saveio/scan/p2p/actor/messages"
	"github.com/saveio/themis/common/log"
)

const (
	REQ_TIMEOUT = 15
)

var DnsP2p *Network

var once sync.Once

const (
	OpCodeTorrent opcode.Opcode = 3000 + iota
	OpCodeEndpoint
)

var opCodes = map[opcode.Opcode]proto.Message{
	OpCodeTorrent:  &pm.Torrent{},
	OpCodeEndpoint: &pm.Endpoint{},
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
	keepalive             *keepalive.Component
}

func NewP2P() *Network {
	n := &Network{
		P2p: new(network.Network),
	}
	n.ActivePeers = new(sync.Map)
	n.addressForHealthCheck = new(sync.Map)
	n.kill = make(chan struct{})
	n.peerStateChan = make(chan *keepalive.PeerStateEvent, 100)
	return n
}

func (this *Network) SetProxyServer(address string) {
	this.proxyAddr = address
}

func (this *Network) SetNetworkKey(key *crypto.KeyPair) {
	this.Keys = key
}

func (this *Network) Protocol() string {
	idx := strings.Index(this.PublicAddr(), "://")
	if idx == -1 {
		return "tcp"
	}
	return this.PublicAddr()[:idx]
}

func (this *Network) Start(address string) error {
	protocolIndex := strings.Index(address, "://")
	if protocolIndex == -1 {
		return errors.New("invalid address")
	}
	protocol := address[:protocolIndex]
	log.Debugf("channel address: %s", address)
	builderOpt := []network.BuilderOption{
		network.WriteFlushLatency(1 * time.Millisecond),
		network.WriteTimeout(int(time.Duration(30))),
	}
	builder := network.NewBuilderWithOptions(builderOpt...)
	if this.Keys != nil {
		log.Debugf("channel use account key")
		builder.SetKeys(this.Keys)
	} else {
		log.Debugf("channel use RandomKeyPair key")
		builder.SetKeys(ed25519.RandomKeyPair())
	}

	builder.SetAddress(address)
	component := new(NetComponent)
	component.Net = this
	builder.AddComponent(component)

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
	this.keepalive = keepalive.New(options...)
	builder.AddComponent(this.keepalive)

	if len(this.proxyAddr) > 0 {
		switch protocol {
		case "udp":
			builder.AddComponent(new(proxy.UDPProxyComponent))
		case "kcp":
			builder.AddComponent(new(proxy.KCPProxyComponent))
		case "quic":
			builder.AddComponent(new(proxy.QuicProxyComponent))
		case "tcp":
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
	this.P2p.SetNetworkID(config.Parameters.Base.DnsNetworkId)
	go this.P2p.Listen()
	// go this.PeerStateChange(this.syncPeerState)

	this.P2p.BlockUntilListening()
	log.Debugf("channel will BlockUntilProxyFinish..., networkid %d", config.Parameters.Base.DnsNetworkId)
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

func (this *Network) Stop() {
	close(this.kill)
	this.P2p.Close()
}

func (this *Network) GetPeerStateByAddress(addr string) (network.PeerState, error) {
	return this.P2p.GetRealConnState(addr)
}

func (this *Network) Connect(tAddr string) error {
	if this == nil {
		return fmt.Errorf("[Connect] this is nil")
	}
	peerState, _ := this.GetPeerStateByAddress(tAddr)
	if peerState == network.PEER_REACHABLE {
		return nil
	}
	if _, ok := this.addressForHealthCheck.Load(tAddr); ok {
		// already try to connect, don't retry before we get a result
		log.Info("already try to connect")
		return nil
	}

	this.addressForHealthCheck.Store(tAddr, struct{}{})
	this.P2p.Bootstrap(tAddr)
	return nil
}

func (this *Network) ConnectAndWait(addr string) error {
	if this.IsConnectionExists(addr) {
		log.Debugf("connection exist %s", addr)
		return nil
	}
	log.Debugf("bootstrap to %v", addr)
	this.P2p.Bootstrap(addr)
	err := this.WaitForConnected(addr, time.Duration(5)*time.Second)
	if err != nil {
		return err
	}
	return nil
}

func (this *Network) WaitForConnected(addr string, timeout time.Duration) error {
	interval := time.Duration(1) * time.Second
	secs := int(timeout / interval)
	if secs <= 0 {
		secs = 1
	}
	for i := 0; i < secs; i++ {
		state, _ := this.GetPeerStateByAddress(addr)
		log.Debugf("this.keepalive: %p, GetPeerStateByAddress state addr:%s, :%d", this.keepalive, addr, state)
		if state == network.PEER_REACHABLE {
			return nil
		}
		<-time.After(interval)
	}
	return errors.New("wait for connected timeout")
}

func (this *Network) IsConnectionExists(addr string) bool {
	if this.P2p == nil {
		return false
	}
	return this.P2p.ConnectionStateExists(addr)
}

// Send send msg to peer asyncnously
// peer can be addr(string) or client(*network.peerClient)
func (this *Network) Send(msg proto.Message, toAddr string) error {
	err := this.healthCheckProxyServer()
	if err != nil {
		return err
	}
	state, _ := this.GetPeerStateByAddress(toAddr)
	if state != network.PEER_REACHABLE {
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

// Request. send msg to peer and wait for response synchronously with timeout
func (this *Network) Request(msg proto.Message, peer string) (proto.Message, error) {
	err := this.healthCheckProxyServer()
	if err != nil {
		return nil, err
	}
	client := this.P2p.GetPeerClient(peer)
	if client == nil {
		return nil, fmt.Errorf("get peer client is nil %s", peer)
	}
	return client.Request(context.Background(), msg, time.Duration(common.REQUEST_MSG_TIMEOUT)*time.Second)
}

// RequestWithRetry. send msg to peer and wait for response synchronously
func (this *Network) RequestWithRetry(msg proto.Message, peer string, retry int) (proto.Message, error) {
	var err error
	err = this.healthCheckProxyServer()
	if err != nil {
		return nil, err
	}
	client := this.P2p.GetPeerClient(peer)
	if client == nil {
		return nil, fmt.Errorf("get peer client is nil %s", peer)
	}
	var res proto.Message
	for i := 0; i < retry; i++ {
		log.Debugf("send request msg to %s with retry %d", peer, i)
		res, err = client.Request(context.Background(), msg, time.Duration(common.REQUEST_MSG_TIMEOUT)*time.Second)
		if err == nil || err.Error() != "context deadline exceeded" {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

//P2P network msg receive. torrent msg, reg msg, unReg msg
func (this *Network) Receive(message proto.Message, from string) error {
	switch message.(type) {
	case *pm.Torrent:
		log.Errorf("[MSB Receive] receive from peer:%s, nil Torrent message", from)
		this.OnBusinessMessage(message, from)
	case *pm.Endpoint:
		log.Errorf("[MSB Receive] receive from peer:%s, nil Reg message", from)
		this.OnBusinessMessage(message, from)
	default:
		// log.Errorf("[MSB Receive] unknown message type:%s", msg.String())
	}
	return nil
}

func (this *Network) OnBusinessMessage(message proto.Message, from string) error {
	log.Debugf("[OnBusinessMessage] receive message from peer:%s", from)
	future := this.GetPID().RequestFuture(message,
		REQ_TIMEOUT*time.Second)
	if _, err := future.Result(); err != nil {
		log.Error("[OnBusinessMessage] error: ", err)
		return err
	}
	return nil
}

func getProtocolFromAddr(addr string) string {
	idx := strings.Index(addr, "://")
	if idx == -1 {
		return "tcp"
	}
	return addr[:idx]
}

func (this *Network) healthCheckProxyServer() error {
	if len(this.proxyAddr) == 0 {
		return nil
	}
	proxyState, err := this.GetPeerStateByAddress(this.proxyAddr)
	if err != nil {
		return err
	}
	if proxyState == network.PEER_REACHABLE {
		return nil
	}
	client := this.P2p.GetPeerClient(this.proxyAddr)
	if client == nil {
		return fmt.Errorf("get peer client is nil: %s", this.proxyAddr)
	}
	proxyState, err = this.GetPeerStateByAddress(this.proxyAddr)
	if err != nil || proxyState != network.PEER_REACHABLE {
		return fmt.Errorf("back off proxy failed state: %d, err: %s", proxyState, err)
	}
	return nil
}
