package tk

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/anacrolix/dht/krpc"
	"github.com/gogo/protobuf/proto"
	"github.com/ontio/ontology-eventbus/actor"
	tkpm "github.com/saveio/scan/p2p/actor/messages"
	tkActClient "github.com/saveio/scan/p2p/actor/tracker/client"
	"github.com/saveio/scan/storage"
	chainsdk "github.com/saveio/themis-go-sdk/utils"
	"github.com/saveio/themis/account"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/crypto/keypair"
)

const (
	MAX_ANNOUNCE_REQUEST_TIMES  = 3
	MAX_ANNOUNCE_RESPONSE_TIMES = 1
	MAX_ANNOUNCE_WAIT_SECONDS   = 30 * time.Second
	MEMORY_HEARTBEAT_DURATION   = 5 * time.Second
)

type ActionEndpointRegParams struct {
	Wallet [20]byte
	IP     net.IP
	Port   uint16
}

type ActionTorrentCompleteParams struct {
	InfoHash storage.MetaInfoHash
	IP       net.IP
	Port     uint16
}

type ActionGetTorrentPeersParams struct {
	InfoHash storage.MetaInfoHash
	NumWant  int32
	Left     uint64
}

type AnnounceAction int32
type AnnounceMessageStatus int32

const (
	AnnounceActionRequest AnnounceAction = iota
	AnnounceActionResponse
)

const (
	AnnounceMessageProcessed AnnounceMessageStatus = iota
)

type AnnounceMessageItem struct {
	MsgID       *tkpm.MessageID
	AnnRequest  *tkpm.AnnounceRequest
	AnnResponse *tkpm.AnnounceResponse
	Action      AnnounceAction             // "request" or "response"
	ChStatus    chan AnnounceMessageStatus // "request processed"、"response delivered"
}

type TrackerService struct {
	AnnounceMessageMap *sync.Map
	TkAct              *actor.PID
	DnsAct             *actor.PID
	PublicKey          keypair.PublicKey
	SignFn             func(rawData []byte) ([]byte, error)
}

func NewTrackerService(tkAct, dnsAct *actor.PID, pubKey keypair.PublicKey, _sigCallback func(rawData []byte) ([]byte, error)) *TrackerService {
	return &TrackerService{
		AnnounceMessageMap: new(sync.Map),
		TkAct:              tkAct,
		DnsAct:             dnsAct,
		PublicKey:          pubKey,
		SignFn:             _sigCallback,
	}
}

func (this *TrackerService) SetTkActor(tkAct *actor.PID) {
	this.TkAct = tkAct
}

func (this *TrackerService) SetDnsActor(dnsAct *actor.PID) {
	this.DnsAct = dnsAct
}

func (this *TrackerService) Start() {
	log.Info("tkSrv started")
	tkActClient.SetTrackerServerPid(this.TkAct)
	for {
		t := time.NewTimer(time.Duration(MEMORY_HEARTBEAT_DURATION))
		select {
		case <-t.C:
			log.Debugf("AnnounceMessageMap: %v", this.AnnounceMessageMap)
			tkActClient.P2pTell()
		}
	}
}

func (this *TrackerService) HandleAnnounceRequestEvent(annReq *tkpm.AnnounceRequest) (*tkpm.AnnounceResponse, error) {
	log.Debugf("HandleAnnounceRequestEvent")
	if annReq.MessageIdentifier == nil {
		annReq.MessageIdentifier = &tkpm.MessageID{
			MessageId: uint64(GetMsgID()),
		}
	}

	annMsgItem := &AnnounceMessageItem{
		MsgID:       annReq.MessageIdentifier,
		AnnRequest:  annReq,
		AnnResponse: nil,
		Action:      AnnounceActionRequest,
		ChStatus:    make(chan AnnounceMessageStatus),
	}

	log.Infof("SignAndSend %v %v %v", annReq.Target, annReq.MessageIdentifier, annReq)
	// this.AnnounceMessageMap.Store(annReq.MessageIdentifier.MessageId, annReq)
	this.SignAndSend(annReq.Target, annReq.MessageIdentifier, annReq)

	var result *tkpm.AnnounceResponse
	var interval time.Duration = MAX_ANNOUNCE_REQUEST_TIMES
	t := time.NewTimer(interval * time.Second)

	// Waiting & Retry
	for {
		select {
		case <-time.After(time.Duration(MAX_ANNOUNCE_WAIT_SECONDS)):
			log.Debugf("Timeout finally")
			msgItem, ok := this.AnnounceMessageMap.Load(annReq.MessageIdentifier)
			if ok && msgItem != nil {
				this.AnnounceMessageMap.Delete(annReq.MessageIdentifier)
			}
			break
		case <-t.C:
			log.Debugf("SignAndSend [QueueSend] <-t.C Time: %s queue: %+v\n", time.Now().String(), annMsgItem.AnnRequest)
			this.SignAndSend(annReq.Target, annReq.MessageIdentifier, annReq)
			log.Warnf("Timeout retry for msg = %+v\n", annReq)
			t.Reset(interval * time.Second)
		case status := <-annMsgItem.ChStatus:
			msgItem_, ok := this.AnnounceMessageMap.Load(annMsgItem.MsgID)
			msgItem := msgItem_.(*AnnounceMessageItem)
			if ok && annMsgItem.MsgID == msgItem.MsgID && status == AnnounceMessageProcessed {
				this.AnnounceMessageMap.Delete(annReq.MessageIdentifier)
				break
			}
		}
	}

	// Sweeper, remove msg from map
	msgItemI, ok := this.AnnounceMessageMap.Load(annReq.MessageIdentifier)
	if ok && msgItemI != nil {
		this.AnnounceMessageMap.Delete(annReq.MessageIdentifier)
	}
	return result, nil
}

func (this *TrackerService) HandleAnnounceResponseEvent(annRes *tkpm.AnnounceResponse, from string) error {
	this.SignAndSend(from, annRes.MessageIdentifier, annRes)

	var interval time.Duration = MAX_ANNOUNCE_RESPONSE_TIMES
	ti := time.NewTimer(interval * time.Second)
	// Waiting & Retry
	for {
		select {
		case <-time.After(time.Duration(MAX_ANNOUNCE_WAIT_SECONDS)):
			log.Debugf("")
		case <-ti.C:
			log.Debugf("[QueueSend] <-t.C Time: %s queue: %+v\n", time.Now().String(), annRes)
			this.SignAndSend(from, annRes.MessageIdentifier, annRes)
			log.Warnf("Timeout retry for msg = %+v\n", annRes)
			ti.Reset(interval * time.Second)
		}
	}
}

func (this *TrackerService) SignAndSend(from string, msgId *tkpm.MessageID, message proto.Message) error {
	params := ActionEndpointRegParams{}
	rawData, err := json.Marshal(params)
	if err != nil {
		return err
	}

	// signature, err := chainsdk.Sign(acc, rawData)
	signature, err := this.SignFn(rawData)
	if err != nil {
		return errors.New("gen message signature failed.")
	}

	switch msg := message.(type) {
	case *tkpm.AnnounceRequest:
		log.Debugf("tkpm.AnnounceRequestMessage Request: %v, Signature: %v", msg, tkpm.SignedMessage{Signature: signature, Publikkey: keypair.SerializePublicKey(this.PublicKey)})
		go tkActClient.P2pSend(from, &tkpm.AnnounceRequestMessage{
			Request: msg,
			Signature: &tkpm.SignedMessage{
				Signature: signature,
				Publikkey: keypair.SerializePublicKey(this.PublicKey),
			},
		})
		break
	case *tkpm.AnnounceResponse:
		go tkActClient.P2pSend(from, &tkpm.AnnounceResponseMessage{
			Response: msg,
			Signature: &tkpm.SignedMessage{
				Signature: signature,
				Publikkey: keypair.SerializePublicKey(this.PublicKey),
			},
		})
		break
	default:
		return fmt.Errorf("Unknown message type to send")
	}
	return nil
}

func (this *TrackerService) CheckSign(acc *account.Account, rawData, signData []byte) error {
	return chainsdk.Verify(acc.PublicKey, rawData, signData)
}

func (this *TrackerService) ReceiveAnnounceMessage(message proto.Message, from string) {
	log.Debug("[NetComponent] Receive: ", reflect.TypeOf(message).String(), " From: ", from)
	switch msg := message.(type) {
	case *tkpm.AnnounceRequestMessage:
		switch msg.GetRequest().GetEvent() {
		case tkpm.AnnounceRequest_EMPTY:
		case tkpm.AnnounceRequest_COMPLETED:
		case tkpm.AnnounceRequest_EMPTY_SEND:
		case tkpm.AnnounceRequest_COMPLETED_SEND:
			annResp, err := this.onAnnounce(msg.GetRequest())
			if err != nil {
				this.HandleAnnounceResponseEvent(&tkpm.AnnounceResponse{}, from)
			}
			this.HandleAnnounceResponseEvent(annResp, from)
		}
	case *tkpm.AnnounceResponseMessage:
		switch msg.GetResponse().GetEvent() {
		case tkpm.AnnounceResponse_EMPTY:
		case tkpm.AnnounceResponse_EMPTY_SEND:
		case tkpm.AnnounceResponse_COMPLETED:
		case tkpm.AnnounceResponse_COMPLETED_SEND:
			this.ReceiveAnnounceResponseMessage(msg, from)
		}
	}
}

func (this *TrackerService) ReceiveAnnounceResponseMessage(annRespMsg *tkpm.AnnounceResponseMessage, from string) {
	msg, ok := this.AnnounceMessageMap.Load(annRespMsg.Response.MessageIdentifier)
	if ok && msg != nil {
		msgItem := msg.(*AnnounceMessageItem)
		msgItem.AnnResponse = annRespMsg.Response
		this.AnnounceMessageMap.Store(annRespMsg.Response.MessageIdentifier.MessageId, msgItem)
		msgItem.ChStatus <- AnnounceMessageProcessed
	}
}

func (this *TrackerService) onAnnounce(arq *tkpm.AnnounceRequest) (*tkpm.AnnounceResponse, error) {
	switch arq.Event {
	case tkpm.AnnounceRequest_EMPTY_SEND:
		return this.onAnnounceQuery(arq)
	case tkpm.AnnounceRequest_COMPLETED_SEND:
		return this.onAnnounceComplete(arq)
	}
	return nil, errors.New("Unknown announce event type")
}

func (this *TrackerService) onAnnounceQuery(arq *tkpm.AnnounceRequest) (*tkpm.AnnounceResponse, error) {
	t, err := storage.TDB.GetTorrent(arq.InfoHash[:])
	if err != nil {
		return nil, err
	}

	var peers []string
	for peer, _ := range t.Peers {
		peers = append(peers, peer)
	}
	return &tkpm.AnnounceResponse{
		MessageIdentifier: arq.MessageIdentifier,
		Interval:          900,
		Leechers:          uint64(t.Leechers),
		Seeders:           uint64(t.Seeders),
		Peers:             peers,
	}, nil
}

func (this *TrackerService) onAnnounceComplete(arq *tkpm.AnnounceRequest) (*tkpm.AnnounceResponse, error) {
	t, err := storage.TDB.GetTorrent(arq.InfoHash)
	if err != nil {
		return nil, err
	}

	peer := krpc.NodeAddr{IP: arq.Ip[:], Port: int(arq.Port)}
	pi, err := storage.TDB.GetTorrentPeerByFileHashAndNodeAddr(arq.InfoHash[:], peer.String())
	if err != nil {
		return nil, err
	}

	if pi == nil {
		var peerID storage.PeerID
		copy(peerID[:storage.PEERID_LENGTH], arq.PeerId[:storage.PEERID_LENGTH])
		pi = &storage.PeerInfo{
			ID:       peerID,
			Complete: arq.Left == 0,
			IP:       ipconvert(net.IP(arq.Ip).To4()),
			Port:     uint16(arq.Port),
			NodeAddr: peer,
		}
		pi.Print()
		err := storage.TDB.AddTorrentPeer(arq.InfoHash[:], arq.Left, peer.String(), pi)
		if err != nil {
			return nil, err
		}
	} else {
		torrent, err := storage.TDB.GetTorrent(arq.InfoHash[:])
		if err != nil {
			return nil, err
		}
		torrent.Seeders += 1
		torrent.Leechers -= 1
		pi.Complete = true
		torrent.Peers[pi.NodeAddr.String()] = pi
		storage.TDB.PutTorrent(arq.InfoHash[:], torrent)
	}

	piBinarys := pi.Serialize()
	log.Debugf("PeerInfo Binary: %v\n", piBinarys)
	this.DnsAct.Tell(&tkpm.Torrent{InfoHash: arq.InfoHash[:], Left: arq.Left, Peerinfo: piBinarys, Type: 0})

	var peers []string
	for peer, _ := range t.Peers {
		peers = append(peers, peer)
	}

	return &tkpm.AnnounceResponse{
		MessageIdentifier: arq.MessageIdentifier,
		Interval:          900,
		Leechers:          uint64(t.Leechers),
		Seeders:           uint64(t.Seeders),
		Peers:             peers,
	}, nil
}

func GetMsgID() int64 {
	for {
		b := new(big.Int).SetInt64(math.MaxInt64)
		if id, err := rand.Int(rand.Reader, b); err == nil {
			messageId := id.Int64()
			return messageId
		}
	}
}

func ipconvert(ip net.IP) (ip_4 [4]byte) {
	if len(ip) != 4 {
		ip = ip.To4()
	}
	copy(ip_4[:4], ip[:4])
	return ip_4
}

// func (this *TrackerService) onAnnounceLeave(arq *tkpm.AnnounceRequest, pi *peerInfo) error {
// 	if pi == nil {
// 		return errors.New("peer not exists.")
// 	}
// 	err := storage.TDB.DelTorrentPeer(arq.InfoHash[:], pi)
// 	if err != nil {
// 		return err
// 	}

// 	piBinarys := pi.Serialize()
// 	pi.Print()
// 	log.Debugf("PeerInfo Binarys: %v\n", piBinarys)
// 	this.dnsAct.p2p.Tell(&pm.Torrent{InfoHash: arq.InfoHash[:], Left: arq.Left, PeerInfo: piBinarys, Type: 0})
// 	return &tkpm.AnnounceResponse{
// 		MessageIdentifier: arq.MessageIdentifier,
// 		Interval: 900,
// 		Leechers: t.Leechers,
// 		Seeders:  t.Seeders,
// 	}, nil
// }

// func (this *TrackerService) MarshalTorrentPeers(t *storage.Torrent) ([]byte, error) {
// 	bm := func() encoding.BinaryMarshaler {
// 		ip := missinggo.AddrIP(addr)
// 		pNodeAddrs := make([]krpc.NodeAddr, 0)
// 		for _, pi := range t.Peers {
// 			if !pi.Complete {
// 				continue
// 			}
// 			nodeAddIp := pi.NodeAddr.IP.To4()
// 			if nodeAddIp == nil {
// 				nodeAddIp = pi.NodeAddr.IP.To16()
// 			}
// 			pNodeAddrs = append(pNodeAddrs, krpc.NodeAddr{
// 				IP:   nodeAddIp,
// 				Port: pi.NodeAddr.Port,
// 			})
// 			if ar.NumWant != -1 && len(pNodeAddrs) >= int(ar.NumWant) {
// 				break
// 			}
// 		}
// 		if ip.To4() != nil {
// 			return krpc.CompactIPv4NodeAddrs(pNodeAddrs)
// 		}
// 		return krpc.CompactIPv6NodeAddrs(pNodeAddrs)
// 	}()
// 	return bm.MarshalBinary()
// }
