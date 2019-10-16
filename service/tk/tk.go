package tk

import (
	"crypto/rand"
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
	theComm "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/crypto/keypair"
)

const (
	MAX_ANNOUNCE_REQUEST_TIMES  = 1
	MAX_ANNOUNCE_RESPONSE_TIMES = 1
	MAX_ANNOUNCE_WAIT_SECONDS   = 30 * time.Second
	MEMORY_HEARTBEAT_DURATION   = 45 * time.Second
	MAX_TORRENT_PEERS_RET_NUM   = 30
)

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
	DnsAct             *actor.PID
	PublicKey          keypair.PublicKey
	SignFn             func(rawData []byte) ([]byte, error)
}

func NewTrackerService(dnsAct *actor.PID, pubKey keypair.PublicKey, _sigCallback func(rawData []byte) ([]byte, error)) *TrackerService {
	return &TrackerService{
		AnnounceMessageMap: new(sync.Map),
		DnsAct:             dnsAct,
		PublicKey:          pubKey,
		SignFn:             _sigCallback,
	}
}

func (this *TrackerService) SetDnsActor(dnsAct *actor.PID) {
	this.DnsAct = dnsAct
}

func (this *TrackerService) Start(targetDnsAddr string) {
	tkActClient.P2pConnect(targetDnsAddr)
	go func() {
		for {
			t := time.NewTimer(time.Duration(MEMORY_HEARTBEAT_DURATION))
			select {
			case <-t.C:
				log.Debugf("AnnounceMessageMap: %v", this.AnnounceMessageMap)
				go tkActClient.P2pConnect(targetDnsAddr)
			}
		}
	}()
	log.Info("tkSrv started")
}

func (this *TrackerService) ConnectDns(targetDnsAddr string) error {
	return tkActClient.P2pConnect(targetDnsAddr)
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
	this.AnnounceMessageMap.Store(annReq.MessageIdentifier.MessageId, annMsgItem)
	go this.SignAndSend(annReq.Target, annReq.MessageIdentifier, annReq)

	var result *tkpm.AnnounceResponse
	var interval time.Duration = 15
	retry_times := 0
	// Waiting & Retry
	for retry_times < MAX_ANNOUNCE_REQUEST_TIMES {
		t := time.NewTimer(interval * time.Second)
		select {
		case <-t.C:
			log.Debugf("SignAndSend [QueueSend] <-t.C Time: %s queue: %+v\n", time.Now().String(), annMsgItem.AnnRequest)
			go this.SignAndSend(annReq.Target, annReq.MessageIdentifier, annReq)
			// log.Warnf("Timeout retry for msg = %+v\n", annReq)
			// t.Reset(interval * time.Second)
			// Sweeper, remove msg from map
			go func() {
				msgItemI, ok := this.AnnounceMessageMap.Load(annReq.MessageIdentifier.MessageId)
				if ok && msgItemI != nil {
					this.AnnounceMessageMap.Delete(annReq.MessageIdentifier.MessageId)
				}
			}()
			retry_times++
			log.Debugf("timeout retry_times %d", retry_times)
		case status := <-annMsgItem.ChStatus:
			log.Infof("gotit")
			retry_times = MAX_ANNOUNCE_REQUEST_TIMES
			msgItem_, ok := this.AnnounceMessageMap.Load(annMsgItem.MsgID.MessageId)
			msgItem := msgItem_.(*AnnounceMessageItem)
			if ok && annMsgItem.MsgID == msgItem.MsgID && status == AnnounceMessageProcessed {
				result = annMsgItem.AnnResponse
				go this.AnnounceMessageMap.Delete(annReq.MessageIdentifier.MessageId)
				return result, nil
			}
		}
	}

	return &tkpm.AnnounceResponse{
		MessageIdentifier: annReq.MessageIdentifier,
		Timeout:           true,
	}, errors.New("timeout")
}

func (this *TrackerService) HandleAnnounceResponseEvent(annRes *tkpm.AnnounceResponse, from string) error {
	go this.SignAndSend(from, annRes.MessageIdentifier, annRes)
	return nil
}

func (this *TrackerService) SignAndSend(target string, msgId *tkpm.MessageID, message proto.Message) (err error) {
	switch msg := message.(type) {
	case *tkpm.AnnounceRequest:
		var rawData []byte
		switch msg.Event {
		case tkpm.AnnounceEvent_COMPLETE_TORRENT:
			rawData, err = proto.Marshal(msg.CompleteTorrentReq)
			if err != nil {
				return err
			}
		case tkpm.AnnounceEvent_QUERY_TORRENT_PEERS:
			rawData, err = proto.Marshal(msg.GetTorrentPeersReq)
			if err != nil {
				return err
			}
		case tkpm.AnnounceEvent_ENDPOINT_REGISTRY:
			rawData, err = proto.Marshal(msg.EndpointRegistryReq)
			if err != nil {
				return err
			}
		case tkpm.AnnounceEvent_QUERY_ENDPOINT:
			rawData, err = proto.Marshal(msg.QueryEndpointReq)
			if err != nil {
				return err
			}
		}
		if len(rawData) == 0 {
			return errors.New("rawData null")
		}

		signature, err := this.SignFn(rawData)
		if err != nil {
			return errors.New("gen message signature failed.")
		}
		log.Debugf("tkpm.AnnounceRequestMessage RawData: %v, PublicKey: %v, Signature: %v", rawData, keypair.SerializePublicKey(this.PublicKey), signature)

		go tkActClient.P2pSend(target, &tkpm.AnnounceRequestMessage{
			Request: msg,
			Signature: &tkpm.SignedMessage{
				Signature: signature,
				Publikkey: keypair.SerializePublicKey(this.PublicKey),
			},
		})
		break
	case *tkpm.AnnounceResponse:
		var rawData []byte
		switch msg.Event {
		case tkpm.AnnounceEvent_COMPLETE_TORRENT:
			rawData, err = proto.Marshal(msg.CompleteTorrentRet)
			if err != nil {
				return err
			}
		case tkpm.AnnounceEvent_QUERY_TORRENT_PEERS:
			rawData, err = proto.Marshal(msg.GetTorrentPeersRet)
			if err != nil {
				return err
			}
		case tkpm.AnnounceEvent_ENDPOINT_REGISTRY:
			rawData, err = proto.Marshal(msg.EndpointRegistryRet)
			if err != nil {
				return err
			}
		case tkpm.AnnounceEvent_QUERY_ENDPOINT:
			rawData, err = proto.Marshal(msg.QueryEndpointRet)
			if err != nil {
				return err
			}
		}
		if len(rawData) == 0 {
			return errors.New("rawData null")
		}

		signature, err := this.SignFn(rawData)
		if err != nil {
			return errors.New("gen message signature failed.")
		}
		log.Debugf("tkpm.AnnounceResponseMessage RawData: %v, PublicKey: %v, Signature: %v", rawData, keypair.SerializePublicKey(this.PublicKey), signature)

		go tkActClient.P2pSend(target, &tkpm.AnnounceResponseMessage{
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

func (this *TrackerService) CheckSign(message proto.Message) (err error) {
	switch msg := message.(type) {
	case *tkpm.AnnounceRequestMessage:
		var rawData []byte
		switch msg.GetRequest().Event {
		case tkpm.AnnounceEvent_COMPLETE_TORRENT:
			rawData, err = proto.Marshal(msg.Request.CompleteTorrentReq)
			if err != nil {
				return err
			}
		case tkpm.AnnounceEvent_QUERY_TORRENT_PEERS:
			rawData, err = proto.Marshal(msg.Request.GetTorrentPeersReq)
			if err != nil {
				return err
			}
		case tkpm.AnnounceEvent_ENDPOINT_REGISTRY:
			rawData, err = proto.Marshal(msg.Request.EndpointRegistryReq)
			if err != nil {
				return err
			}
		case tkpm.AnnounceEvent_QUERY_ENDPOINT:
			rawData, err = proto.Marshal(msg.Request.QueryEndpointReq)
			if err != nil {
				return err
			}
		}
		if len(rawData) == 0 {
			return errors.New("rawData null")
		}

		pubKey, err := keypair.DeserializePublicKey(msg.Signature.Publikkey)
		if err != nil {
			return err
		}
		log.Debugf("tkpm.AnnounceRequestMessage RawData: %v, PublicKey: %v, Signature: %v", rawData, pubKey, msg.Signature.Signature)
		return chainsdk.Verify(pubKey, rawData, msg.Signature.Signature)
	case *tkpm.AnnounceResponseMessage:
		var rawData []byte
		switch msg.GetResponse().Event {
		case tkpm.AnnounceEvent_COMPLETE_TORRENT:
			rawData, err = proto.Marshal(msg.Response.CompleteTorrentRet)
			if err != nil {
				return err
			}
		case tkpm.AnnounceEvent_QUERY_TORRENT_PEERS:
			rawData, err = proto.Marshal(msg.Response.GetTorrentPeersRet)
			if err != nil {
				return err
			}
		case tkpm.AnnounceEvent_ENDPOINT_REGISTRY:
			rawData, err = proto.Marshal(msg.Response.EndpointRegistryRet)
			if err != nil {
				return err
			}
		case tkpm.AnnounceEvent_QUERY_ENDPOINT:
			rawData, err = proto.Marshal(msg.Response.QueryEndpointRet)
			if err != nil {
				return err
			}
		}
		if len(rawData) == 0 {
			return errors.New("rawData null")
		}
		pubKey, err := keypair.DeserializePublicKey(msg.Signature.Publikkey)
		if err != nil {
			return err
		}
		log.Debugf("tkpm.AnnounceResponseMessage RawData: %v, PublicKey: %v, Signature: %v", rawData, pubKey, msg.Signature.Signature)
		return chainsdk.Verify(pubKey, rawData, msg.Signature.Signature)
	}
	return errors.New("unkonwn message type to CheckSign")
}

func (this *TrackerService) ReceiveAnnounceMessage(message proto.Message, from string) {
	log.Debug("[NetComponent] Receive: ", reflect.TypeOf(message).String(), " From: ", from)
	if err := this.CheckSign(message); err != nil {
		log.Errorf("CheckSign failed: %v", err)
		return
	}
	switch msg := message.(type) {
	case *tkpm.AnnounceRequestMessage:
		log.Debugf("msg type %s", msg.GetRequest().Event.String())
		annResp, err := this.onAnnounce(msg.GetRequest())
		log.Debugf("onAnnounce annResp: %v, Err: %v", annResp, err)

		if err != nil {
			switch msg.GetRequest().GetEvent() {
			case tkpm.AnnounceEvent_COMPLETE_TORRENT:
				log.Debugf("AnnounceEvent_COMPLETE_TORRENT")
				this.HandleAnnounceResponseEvent(&tkpm.AnnounceResponse{
					MessageIdentifier: msg.Request.MessageIdentifier,
					CompleteTorrentRet: &tkpm.CompleteTorrentRet{
						Status: tkpm.AnnounceRetStatus_FAIL,
						ErrMsg: err.Error(),
					},
					Event: msg.Request.Event,
				}, from)
			case tkpm.AnnounceEvent_QUERY_TORRENT_PEERS:
				log.Debugf("AnnounceEvent_QUERY_TORRENT_PEERS")
				this.HandleAnnounceResponseEvent(&tkpm.AnnounceResponse{
					MessageIdentifier: msg.Request.MessageIdentifier,
					GetTorrentPeersRet: &tkpm.GetTorrentPeersRet{
						Status: tkpm.AnnounceRetStatus_FAIL,
						ErrMsg: err.Error(),
					},
					Event: msg.Request.Event,
				}, from)
			case tkpm.AnnounceEvent_ENDPOINT_REGISTRY:
				log.Debugf("AnnounceEvent_ENDPOINT_REGISTRY")
				this.HandleAnnounceResponseEvent(&tkpm.AnnounceResponse{
					MessageIdentifier: msg.Request.MessageIdentifier,
					EndpointRegistryRet: &tkpm.EndpointRegistryRet{
						Status: tkpm.AnnounceRetStatus_FAIL,
						ErrMsg: err.Error(),
					},
					Event: msg.Request.Event,
				}, from)
			case tkpm.AnnounceEvent_QUERY_ENDPOINT:
				log.Debugf("AnnounceEvent_QUERY_ENDPOINT")
				this.HandleAnnounceResponseEvent(&tkpm.AnnounceResponse{
					MessageIdentifier: msg.Request.MessageIdentifier,
					QueryEndpointRet: &tkpm.QueryEndpointRet{
						Status: tkpm.AnnounceRetStatus_FAIL,
						ErrMsg: err.Error(),
					},
					Event: msg.Request.Event,
				}, from)
			default:
				log.Debugf("Unknown AnnounceEvent type")
			}
		} else {
			this.HandleAnnounceResponseEvent(annResp, from)
		}
	case *tkpm.AnnounceResponseMessage:
		this.ReceiveAnnounceResponseMessage(msg, from)
		switch msg.GetResponse().GetEvent() {
		case tkpm.AnnounceEvent_COMPLETE_TORRENT:
			log.Debugf("AnnounceEvent_COMPLETE_TORRENT")
		case tkpm.AnnounceEvent_QUERY_TORRENT_PEERS:
			log.Debugf("AnnounceEvent_QUERY_TORRENT_PEERS")
		case tkpm.AnnounceEvent_ENDPOINT_REGISTRY:
			log.Debugf("AnnounceEvent_ENDPOINT_REGISTRY")
		case tkpm.AnnounceEvent_QUERY_ENDPOINT:
			log.Debugf("AnnounceEvent_QUERY_ENDPOINT")
		default:
			log.Debugf("Unknown AnnounceEvent type")
		}
		this.ReceiveAnnounceResponseMessage(msg, from)
	}
}

func (this *TrackerService) ReceiveAnnounceResponseMessage(annRespMsg *tkpm.AnnounceResponseMessage, from string) {
	log.Infof("ReceiveAnnounceResponseMessage %v", annRespMsg)
	msg, ok := this.AnnounceMessageMap.Load(annRespMsg.Response.MessageIdentifier.MessageId)
	if ok && msg != nil {
		log.Debugf("msg: %v", msg)
		msgItem := msg.(*AnnounceMessageItem)
		msgItem.AnnResponse = annRespMsg.Response
		this.AnnounceMessageMap.Store(annRespMsg.Response.MessageIdentifier.MessageId, msgItem)
		msgItem.ChStatus <- AnnounceMessageProcessed
	}
}

func (this *TrackerService) onAnnounce(aReq *tkpm.AnnounceRequest) (*tkpm.AnnounceResponse, error) {
	switch aReq.Event {
	case tkpm.AnnounceEvent_COMPLETE_TORRENT:
		return this.onAnnounceCompleteTorrent(aReq)
	case tkpm.AnnounceEvent_QUERY_TORRENT_PEERS:
		return this.onAnnounceQueryTorrentPeers(aReq)
	case tkpm.AnnounceEvent_ENDPOINT_REGISTRY:
		return this.onAnnounceEndpointRegistry(aReq)
	case tkpm.AnnounceEvent_QUERY_ENDPOINT:
		return this.onAnnounceQueryEndpoint(aReq)
	}
	return nil, errors.New("Unknown announce event type")
}

func (this *TrackerService) onAnnounceEndpointRegistry(aReq *tkpm.AnnounceRequest) (*tkpm.AnnounceResponse, error) {
	req := aReq.EndpointRegistryReq
	if req == nil {
		return nil, errors.New("request.EndpointRegistryReq is nil")
	}
	if storage.EDB == nil {
		return nil, errors.New("storage.EDB is nil")
	}
	var wallet theComm.Address
	copy(wallet[:], req.Wallet)
	peer := krpc.NodeAddr{IP: req.Ip[:], Port: int(req.Port)}
	err := storage.EDB.PutEndpoint(wallet.ToBase58(), peer.IP, int(req.Port))
	if err != nil {
		return nil, err
	}

	if this.DnsAct != nil {
		this.DnsAct.Tell(&tkpm.Endpoint{WalletAddr: wallet.ToBase58(), HostPort: peer.String(), Type: 0})
	}

	return &tkpm.AnnounceResponse{
		MessageIdentifier: aReq.MessageIdentifier,
		EndpointRegistryRet: &tkpm.EndpointRegistryRet{
			Status: tkpm.AnnounceRetStatus_SUCCESS,
			ErrMsg: "",
		},
		Event: aReq.Event,
	}, nil
}

func (this *TrackerService) onAnnounceQueryEndpoint(aReq *tkpm.AnnounceRequest) (*tkpm.AnnounceResponse, error) {
	req := aReq.QueryEndpointReq
	if req == nil {
		return nil, errors.New("request.QueryEndpointReq is nil")
	}
	if storage.EDB == nil {
		return nil, errors.New("storage.EDB is nil")
	}
	var wallet theComm.Address
	copy(wallet[:], req.Wallet)
	if wallet == theComm.ADDRESS_EMPTY {
		return nil, errors.New("wallet invalid, ADDRESS_EMPTY")
	}
	var peer *storage.Endpoint
	peer, err := storage.EDB.GetEndpoint(wallet.ToBase58())
	if err != nil {
		return nil, err
	}
	return &tkpm.AnnounceResponse{
		MessageIdentifier: aReq.MessageIdentifier,
		QueryEndpointRet: &tkpm.QueryEndpointRet{
			Status: tkpm.AnnounceRetStatus_SUCCESS,
			Peer:   peer.NodeAddr.String(),
			ErrMsg: "",
		},
		Event: aReq.Event,
	}, nil

}

func (this *TrackerService) onAnnounceQueryTorrentPeers(aReq *tkpm.AnnounceRequest) (*tkpm.AnnounceResponse, error) {
	req := aReq.GetTorrentPeersReq
	if req == nil {
		return nil, errors.New("request.GetTorrentPeersReq is nil")
	}
	if storage.TDB == nil {
		return nil, errors.New("storage.TDB is nil")
	}
	pis, _, _, err := storage.TDB.GetTorrentPeersByFileHash(req.InfoHash, int32(req.NumWant))
	if err != nil && err.Error() != "not found" {
		return nil, err
	}

	var peers []string
	for _, peer := range pis {
		if len(peers) >= MAX_TORRENT_PEERS_RET_NUM {
			break
		}
		peers = append(peers, peer.NodeAddr.String())
	}
	log.Debugf("onAnnounceQueryTorrentPeers %v", peers)
	return &tkpm.AnnounceResponse{
		MessageIdentifier: aReq.MessageIdentifier,
		GetTorrentPeersRet: &tkpm.GetTorrentPeersRet{
			Status: tkpm.AnnounceRetStatus_SUCCESS,
			Peers:  peers,
			ErrMsg: "",
		},
		Event: aReq.Event,
	}, nil
}

func (this *TrackerService) onAnnounceCompleteTorrent(aReq *tkpm.AnnounceRequest) (*tkpm.AnnounceResponse, error) {
	req := aReq.CompleteTorrentReq
	if req == nil {
		return nil, errors.New("request.CompleteTorrentReq is nil")
	}
	if storage.TDB == nil {
		return nil, errors.New("storage.TDB is nil")
	}
	peer := krpc.NodeAddr{IP: req.Ip[:], Port: int(req.Port)}
	pi, err := storage.TDB.GetTorrentPeerByFileHashAndNodeAddr(req.InfoHash[:], peer.String())
	if err != nil {
		log.Errorf("get peerinfo err: %v", err)
	}

	if pi == nil {
		var peerID storage.PeerID
		rand.Read(peerID[:])
		pi = &storage.PeerInfo{
			ID:       peerID,
			Complete: true,
			IP:       ipconvert(net.IP(req.Ip).To4()),
			Port:     uint16(req.Port),
			NodeAddr: peer,
		}
		pi.Print()
		err := storage.TDB.AddTorrentPeer(req.InfoHash, 0, peer.String(), pi)
		if err != nil {
			return nil, err
		}
	} else {
		torrent, err := storage.TDB.GetTorrent(req.InfoHash[:])
		if err != nil {
			return nil, err
		}
		torrent.Seeders += 1
		torrent.Leechers -= 1
		pi.Complete = true
		torrent.Peers[pi.NodeAddr.String()] = pi
		storage.TDB.PutTorrent(req.InfoHash, torrent)
	}

	piBinarys := pi.Serialize()
	log.Debugf("PeerInfo Binary: %v\n", piBinarys)
	if this.DnsAct != nil {
		this.DnsAct.Tell(&tkpm.Torrent{InfoHash: req.InfoHash, Left: 0, Peerinfo: piBinarys, Type: 0})
	}

	t, err := storage.TDB.GetTorrent(req.InfoHash)
	if err != nil && err.Error() != "not found" {
		return nil, err
	}
	var peers []string
	for peer, _ := range t.Peers {
		peers = append(peers, peer)
	}

	log.Debugf("ret %v", tkpm.AnnounceResponse{
		MessageIdentifier: aReq.MessageIdentifier,
		CompleteTorrentRet: &tkpm.CompleteTorrentRet{
			Status: tkpm.AnnounceRetStatus_SUCCESS,
			Peer:   peer.String(),
			ErrMsg: "",
		},
		Event: aReq.Event,
	})

	return &tkpm.AnnounceResponse{
		MessageIdentifier: aReq.MessageIdentifier,
		CompleteTorrentRet: &tkpm.CompleteTorrentRet{
			Status: tkpm.AnnounceRetStatus_SUCCESS,
			Peer:   peer.String(),
			ErrMsg: "",
		},
		Event: aReq.Event,
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
