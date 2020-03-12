package network

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/carrier/crypto"
	"github.com/saveio/carrier/crypto/ed25519"
	"github.com/saveio/carrier/network"
	"github.com/saveio/carrier/network/components/ackreply"
	"github.com/saveio/carrier/network/components/keepalive"
	"github.com/saveio/carrier/types/opcode"
	"github.com/saveio/dsp-go-sdk/network/common"
	act "github.com/saveio/pylons/actor/server"
	"github.com/saveio/pylons/network/transport/messages"
	scanCom "github.com/saveio/scan/common"
	pm "github.com/saveio/scan/p2p/actor/messages"
	chainCom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

const (
	REQ_TIMEOUT         = 15
	REQUEST_MSG_TIMEOUT = 30
)

var DDNSP2P *Network
var DnsP2p *Network
var TkP2p *Network

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
)

const (
	OpCodeTorrent opcode.Opcode = 3000 + iota
	OpCodeEndpoint
)

const (
	OpCodeAnnounceRequestMessage opcode.Opcode = 4000 + iota
	OpCodeAnnounceResponseMessage
)

var ChannelOpCodes = map[opcode.Opcode]proto.Message{
	OpCodeProcessed:                &messages.Processed{},
	OpCodeDelivered:                &messages.Delivered{},
	OpCodeSecretRequest:            &messages.SecretRequest{},
	OpCodeRevealSecret:             &messages.RevealSecret{},
	OpCodeSecretMsg:                &messages.BalanceProof{},
	OpCodeDirectTransfer:           &messages.DirectTransfer{},
	OpCodeLockedTransfer:           &messages.LockedTransfer{},
	OpCodeRefundTransfer:           &messages.RefundTransfer{},
	OpCodeLockExpired:              &messages.LockExpired{},
	OpCodeWithdrawRequest:          &messages.WithdrawRequest{},
	OpCodeWithdraw:                 &messages.Withdraw{},
	OpCodeCooperativeSettleRequest: &messages.CooperativeSettleRequest{},
	OpCodeCooperativeSettle:        &messages.CooperativeSettle{},
}

var TorrentOpCodes = map[opcode.Opcode]proto.Message{
	OpCodeTorrent:  &pm.Torrent{},
	OpCodeEndpoint: &pm.Endpoint{},
}

var TrackerOpCodes = map[opcode.Opcode]proto.Message{
	OpCodeAnnounceRequestMessage:  &pm.AnnounceRequestMessage{},
	OpCodeAnnounceResponseMessage: &pm.AnnounceResponseMessage{},
}

type Network struct {
	P2p                   *network.Network
	networkId             uint32
	listenAddr            string
	intranetIP            string
	pid                   *actor.PID
	opCodes               map[opcode.Opcode]proto.Message
	protocol              string
	address               string
	keys                  *crypto.KeyPair
	kill                  chan struct{}
	addressForHealthCheck *sync.Map
	walletAddrFromPeerId  func(string) string // peer key from id delegate
	peers                 *sync.Map
}

func NewP2P(opts ...NetworkOption) *Network {
	n := &Network{
		P2p:                   new(network.Network),
		addressForHealthCheck: new(sync.Map),
		kill:                  make(chan struct{}),
		intranetIP:            "127.0.0.1",
		peers:                 new(sync.Map),
	}
	for _, opt := range opts {
		opt.apply(n)
	}
	return n
}

func (this *Network) Protocol() string {
	idx := strings.Index(this.PublicAddr(), "://")
	if idx == -1 {
		return "tcp"
	}
	return this.PublicAddr()[:idx]
}

func (this *Network) Start(protocol, addr, port string) error {
	builderOpt := []network.BuilderOption{
		network.WriteFlushLatency(1 * time.Millisecond),
		network.WriteTimeout(int(time.Duration(30))),
	}
	builder := network.NewBuilderWithOptions(builderOpt...)
	if this.keys == nil {
		this.keys = ed25519.RandomKeyPair()
		log.Debugf("radom")
	}
	log.Debugf("set key %x, %x", this.keys.PublicKey, this.keys.PrivateKey)
	builder.SetKeys(this.keys)
	builder.SetListenAddr(fmt.Sprintf("%s://%s:%s", protocol, this.intranetIP, port))
	builder.SetAddress(fmt.Sprintf("%s://%s:%s", protocol, addr, port))
	log.Debugf("network key %x, intranet ip %s, start at %s, listen at %s",
		this.keys.PublicKey, this.intranetIP, fmt.Sprintf("%s://%s:%s", protocol, this.intranetIP, port),
		fmt.Sprintf("%s://%s:%s", protocol, addr, port))

	// add msg component
	component := new(NetComponent)
	component.Net = this
	builder.AddComponent(component)

	// add peer component
	peerComp := new(PeerComponent)
	peerComp.Net = this
	builder.AddComponent(peerComp)

	options := []keepalive.ComponentOption{
		keepalive.WithKeepaliveInterval(keepalive.DefaultKeepaliveInterval),
		keepalive.WithKeepaliveTimeout(keepalive.DefaultKeepaliveTimeout),
	}

	builder.AddComponent(keepalive.New(options...))

	// add ack reply
	ackOption := []ackreply.ComponentOption{
		ackreply.WithAckCheckedInterval(time.Second * scanCom.ACK_MSG_CHECK_INTERVAL),
		ackreply.WithAckMessageTimeout(time.Second * scanCom.MAX_ACK_MSG_TIMEOUT),
	}
	builder.AddComponent(ackreply.New(ackOption...))
	var err error
	this.P2p, err = builder.Build()
	if err != nil {
		log.Error("[P2pNetwork] Start builder.Build error: ", err.Error())
		return err
	}
	var once sync.Once
	once.Do(func() {
		for k, v := range this.opCodes {
			if err := opcode.RegisterMessageType(k, v); err != nil {
				log.Errorf("register msg type %v err", k)
			}
			log.Debugf("RegisterMessageType %d", k)
		}
	})
	this.P2p.SetNetworkID(this.networkId)
	this.P2p.SetBootstrapWaitSecond(time.Duration(15) * time.Second)
	go this.P2p.Listen()

	this.P2p.BlockUntilListening()
	log.Debugf("channel will BlockUntilProxyFinish..., networkid %d", this.networkId)
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

func (this *Network) Stop() {
	close(this.kill)
	this.P2p.Close()
	log.Debug("scan network sotp")
}

func (this *Network) GetPeerIdFromWallet(walletAddr string) string {
	p, ok := this.peers.Load(walletAddr)
	if !ok {
		return ""
	}
	pr, ok := p.(*Peer)
	if !ok {
		return ""
	}
	return pr.GetPeerId()
}

func (this *Network) GetPeerStateByAddress(walletAddr string) (network.PeerState, error) {
	if strings.Contains(walletAddr, "tcp") || len(walletAddr) == 0 {
		log.Errorf("wallet addr is wrong %s", walletAddr)
		panic("wallet addr is wrong ")
		os.Exit(1)
	}
	peerId := this.GetPeerIdFromWallet(walletAddr)
	log.Debugf("get peer id from wallet %s %s", walletAddr, peerId)
	if len(peerId) == 0 {
		return 0, fmt.Errorf("peer %s not found", walletAddr)
	}
	return this.P2p.GetRealConnState(peerId)
}

func (this *Network) Connect(hostAddr string) (string, error) {
	if this == nil {
		return "", fmt.Errorf("network is nil")
	}
	log.Debugf("connect to %s", hostAddr)
	walletAddr := this.GetWalletFromHostAddr(hostAddr)
	peer := New(hostAddr)
	if len(walletAddr) > 0 {
		if this.IsConnReachable(walletAddr) {
			log.Debugf("get wallet %s from host %s", hostAddr, walletAddr)
			return walletAddr, nil
		}
		p, ok := this.peers.Load(walletAddr)
		if ok {
			peer, ok = p.(*Peer)
			if !ok {
				peer = New(hostAddr)
			}
		}
	}
	_, ok := this.addressForHealthCheck.LoadOrStore(hostAddr, struct{}{})
	if ok {
		return this.waitForConnectedByHostAddr(hostAddr, time.Duration(15)*time.Second)
	}
	log.Debugf("bootstrap to %v ...", hostAddr)
	peerIds := this.P2p.Bootstrap([]string{hostAddr})
	this.addressForHealthCheck.Delete(hostAddr)
	if len(peerIds) == 0 {
		log.Errorf("bootstrap to %s, no peer ids", hostAddr)
		return "", fmt.Errorf("no peer id for addr %s", hostAddr)
	}
	peerId := peerIds[0]
	walletAddr = this.walletAddrFromPeerId(peerId)
	log.Debugf("host %s, peer id %s, wallet %s", hostAddr, peerId, walletAddr)
	peer.SetPeerId(peerId)
	log.Debugf("Store wallet %s for peer %s %s", walletAddr, hostAddr, peerId)
	this.peers.Store(walletAddr, peer)
	this.addressForHealthCheck.Store(walletAddr, hostAddr)
	return walletAddr, nil
}

func (this *Network) waitForConnectedByHostAddr(hostAddr string, timeout time.Duration) (string, error) {
	interval := time.Duration(1) * time.Second
	secs := int(timeout / interval)
	if secs <= 0 {
		secs = 1
	}
	walletAddr := ""
	for i := 0; i < secs; i++ {
		this.addressForHealthCheck.Range(func(k, v interface{}) bool {
			host, _ := v.(string)
			if host != hostAddr {
				return true
			}
			wallet, ok := k.(string)
			if !ok {
				return true
			}
			walletAddr = wallet
			return false
		})
		if len(walletAddr) > 0 {
			break
		}
		<-time.After(interval)
	}
	if len(walletAddr) == 0 {
		log.Errorf("wait for connect %s timeout", hostAddr)
		return "", errors.New("wait for connected timeout")
	}
	log.Debugf("wait for connect %s success", hostAddr)
	return walletAddr, nil
}

func (this *Network) WaitForConnected(walletAddr string, timeout time.Duration) error {
	if strings.Contains(walletAddr, "tcp") || len(walletAddr) == 0 {
		log.Errorf("wallet addr is wrong %s", walletAddr)
		panic("wallet addr is wrong ")
		os.Exit(1)
	}
	interval := time.Duration(1) * time.Second
	secs := int(timeout / interval)
	if secs <= 0 {
		secs = 1
	}
	for i := 0; i < secs; i++ {
		state, _ := this.GetPeerStateByAddress(walletAddr)
		log.Debugf(" GetPeerStateByAddress state addr:%s, :%d", walletAddr, state)
		if state == network.PEER_REACHABLE {
			return nil
		}
		<-time.After(interval)
	}
	return errors.New("wait for connected timeout")
}

// Send send msg to peer asyncnously
// peer can be addr(string) or client(*network.peerClient)
func (this *Network) Send(msg proto.Message, walletAddr string) error {
	if strings.Contains(walletAddr, "tcp") || len(walletAddr) == 0 {
		log.Errorf("wallet addr is wrong %s", walletAddr)
		panic("wallet addr is wrong ")
		os.Exit(1)
	}
	state, _ := this.GetPeerStateByAddress(walletAddr)
	log.Debugf("get peer state %d", state)
	if state != network.PEER_REACHABLE {
		return fmt.Errorf("can not send to inactive peer %s", walletAddr)
	}
	signed, err := this.P2p.PrepareMessage(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to sign message %s", err)
	}
	log.Debugf("+++ write msg to %s", walletAddr)
	peerId := this.GetPeerIdFromWallet(walletAddr)
	err = this.P2p.Write(peerId, signed)
	log.Debugf("+++ write msg done to %s", walletAddr)
	if err != nil {
		return fmt.Errorf("failed to send message to %s", walletAddr)
	}
	return nil
}

func (this *Network) ListenAddr() string {
	return this.listenAddr
}

func (this *Network) PublicAddr() string {
	return this.P2p.ID.Address
}

// GetPID returns p2p actor
func (this *Network) GetPID() *actor.PID {
	return this.pid
}

// Request. send msg to peer and wait for response synchronously with timeout
func (this *Network) Request(msg proto.Message, walletAddr string) (proto.Message, error) {
	if strings.Contains(walletAddr, "tcp") || len(walletAddr) == 0 {
		log.Errorf("wallet addr is wrong %s", walletAddr)
		panic("wallet addr is wrong ")
		os.Exit(1)
	}
	peerId := this.GetPeerIdFromWallet(walletAddr)
	client := this.P2p.GetPeerClient(peerId)
	if client == nil {
		return nil, fmt.Errorf("get peer client is nil %s", walletAddr)
	}
	return client.Request(context.Background(), msg, time.Duration(common.REQUEST_MSG_TIMEOUT)*time.Second)
}

// RequestWithRetry. send msg to peer and wait for response synchronously
func (this *Network) RequestWithRetry(msg proto.Message, walletAddr string, retry int) (proto.Message, error) {
	if strings.Contains(walletAddr, "tcp") || len(walletAddr) == 0 {
		log.Errorf("wallet addr is wrong %s", walletAddr)
		panic("wallet addr is wrong ")
		os.Exit(1)
	}
	var err error
	peerId := this.GetPeerIdFromWallet(walletAddr)
	client := this.P2p.GetPeerClient(peerId)
	if client == nil {
		return nil, fmt.Errorf("get peer client is nil %s", peerId)
	}
	var res proto.Message
	for i := 0; i < retry; i++ {
		log.Debugf("send request msg to %s with retry %d", peerId, i)
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
func (this *Network) Receive(message proto.Message, from, peerId string) error {
	walletAddr := this.walletAddrFromPeerId(peerId)
	log.Debugf("receive msg from peer %s, wallet %s", peerId, walletAddr)
	switch msg := message.(type) {
	case *messages.Processed:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.Delivered:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.SecretRequest:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.RevealSecret:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.BalanceProof:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.DirectTransfer:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.LockedTransfer:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.RefundTransfer:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.LockExpired:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.WithdrawRequest:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.Withdraw:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.CooperativeSettleRequest:
		act.OnBusinessMessage(message, walletAddr)
	case *messages.CooperativeSettle:
		act.OnBusinessMessage(message, walletAddr)
	case *pm.Torrent:
		log.Debugf("[MSB Receive] receive Torrent message from peer: %s.", walletAddr)
		this.GetPID().Tell(msg)
	case *pm.Endpoint:
		log.Debugf("[MSB Receive] receive Endpoint message from peer: %s.", walletAddr)
		this.GetPID().Tell(msg)
	case *pm.AnnounceRequestMessage:
		log.Debugf("[MSB Receive] receive AnnounceRequestMessage message from peer:%s.", walletAddr)
		msg.GetRequest().From = walletAddr
		this.GetPID().Tell(msg)
	case *pm.AnnounceResponseMessage:
		log.Debugf("[MSB Receive] receive AnnounceResponseMessage message from peer:%s.", walletAddr)
		msg.GetResponse().From = walletAddr
		this.GetPID().Tell(msg)
	default:
		// log.Errorf("[MSB Receive] unknown message type:%s", msg.String())
	}
	return nil
}

func (this *Network) GetWalletFromHostAddr(hostAddr string) string {
	walletAddr := ""
	this.peers.Range(func(key, value interface{}) bool {
		pr, ok := value.(*Peer)
		if pr == nil || !ok {
			return true
		}
		if pr.GetHostAddr() == hostAddr {
			walletAddr = key.(string)
			return false
		}
		return true
	})
	return walletAddr
}

// IsConnReachable. check if peer state reachable
func (this *Network) IsConnReachable(walletAddr string) bool {
	if !isValidWalletAddr(walletAddr) {
		log.Errorf("wrong wallet address [%s]", debug.Stack())
	}
	if this.P2p == nil || len(walletAddr) == 0 {
		return false
	}
	state, err := this.GetPeerStateByAddress(walletAddr)
	log.Debugf("get peer state by addr: %s %v %s", walletAddr, state, err)
	if err != nil {
		return false
	}
	if state != network.PEER_REACHABLE {
		return false
	}
	return true
}

func getProtocolFromAddr(addr string) string {
	idx := strings.Index(addr, "://")
	if idx == -1 {
		return "tcp"
	}
	return addr[:idx]
}

func isValidWalletAddr(walletAddr string) bool {
	_, err := chainCom.AddressFromBase58(walletAddr)
	if err != nil {
		return false
	}
	return true
}
