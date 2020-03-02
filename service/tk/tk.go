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
	"github.com/saveio/themis/core/types"
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
	MsgID       *tkpm.MsgID
	AnnRequest  *tkpm.AnnounceRequest
	AnnResponse *tkpm.AnnounceResponse
	Action      AnnounceAction             // "request" or "response"
	ChStatus    chan AnnounceMessageStatus // "request processed"、"response delivered"
}

type CryptoIdentity struct {
	PubKey  keypair.PublicKey
	RawData []byte
	Sign    []byte
}

type TrackerService struct {
	AnnounceMessageMap *sync.Map
	DnsAct             *actor.PID
	PublicKey          keypair.PublicKey
	SignFn             func(rawData []byte) ([]byte, error)
	CheckWhiteListFn   func(fileHash string, WallAddr theComm.Address) (bool, error)
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

func (this *TrackerService) SetCheckWhiteListFn(_getWhiteListAddrsCallback func(string, theComm.Address) (bool, error)) {
	this.CheckWhiteListFn = _getWhiteListAddrsCallback
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
	_, err := tkActClient.P2pConnect(targetDnsAddr)
	return err
}

func (this *TrackerService) HandleAnnounceRequestEvent(annReq *tkpm.AnnounceRequest) (*tkpm.AnnounceResponse, error) {
	log.Debugf("HandleAnnounceRequestEvent")
	if annReq.MessageIdentifier == nil {
		annReq.MessageIdentifier = &tkpm.MsgID{
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
	go func() {
		err := this.SignAndSend(from, annRes.MessageIdentifier, annRes)
		if err != nil {
			log.Errorf("sign annd send err %s", err)
		}
	}()
	return nil
}

func (this *TrackerService) SignAndSend(target string, msgId *tkpm.MsgID, message proto.Message) (err error) {
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
			log.Errorf("raw data is nil")
			return errors.New("rawData null")
		}

		signature, err := this.SignFn(rawData)
		if err != nil {
			log.Errorf("raw data is nil")
			return errors.New("gen message signature failed.")
		}
		log.Debugf("tkpm.AnnounceRequestMessage RawData: %v, PublicKey: %v, Signature: %v,  target: %v",
			rawData, keypair.SerializePublicKey(this.PublicKey), signature, target)

		go tkActClient.P2pSend(target, &tkpm.AnnounceRequestMessage{
			Request: msg,
			Signature: &tkpm.SignedMsg{
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
		log.Debugf("target %v", target)

		go tkActClient.P2pSend(target, &tkpm.AnnounceResponseMessage{
			Response: msg,
			Signature: &tkpm.SignedMsg{
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

func (this *TrackerService) GetCryptoInfo(message proto.Message) (CryptoIdentity, error) {
	var err error
	var id CryptoIdentity
	var pubKeyBuf []byte

	switch msg := message.(type) {
	case *tkpm.AnnounceRequestMessage:
		switch msg.GetRequest().Event {
		case tkpm.AnnounceEvent_COMPLETE_TORRENT:
			id.RawData, err = proto.Marshal(msg.Request.CompleteTorrentReq)
		case tkpm.AnnounceEvent_QUERY_TORRENT_PEERS:
			id.RawData, err = proto.Marshal(msg.Request.GetTorrentPeersReq)
		case tkpm.AnnounceEvent_ENDPOINT_REGISTRY:
			id.RawData, err = proto.Marshal(msg.Request.EndpointRegistryReq)
		case tkpm.AnnounceEvent_QUERY_ENDPOINT:
			id.RawData, err = proto.Marshal(msg.Request.QueryEndpointReq)
		}
		id.Sign = msg.Signature.Signature
		pubKeyBuf = msg.Signature.Publikkey
	case *tkpm.AnnounceResponseMessage:
		switch msg.GetResponse().Event {
		case tkpm.AnnounceEvent_COMPLETE_TORRENT:
			id.RawData, err = proto.Marshal(msg.Response.CompleteTorrentRet)
		case tkpm.AnnounceEvent_QUERY_TORRENT_PEERS:
			id.RawData, err = proto.Marshal(msg.Response.GetTorrentPeersRet)
		case tkpm.AnnounceEvent_ENDPOINT_REGISTRY:
			id.RawData, err = proto.Marshal(msg.Response.EndpointRegistryRet)
		case tkpm.AnnounceEvent_QUERY_ENDPOINT:
			id.RawData, err = proto.Marshal(msg.Response.QueryEndpointRet)
		}
		id.Sign = msg.Signature.Signature
		pubKeyBuf = msg.Signature.Publikkey
	}
	log.Debugf("get crypto info id: %v, err: %v", id, err)

	if err != nil {
		return id, err
	}
	if len(id.RawData) == 0 {
		return id, errors.New("raw data nil")
	}

	id.PubKey, err = keypair.DeserializePublicKey(pubKeyBuf)
	if err != nil {
		return id, err
	}
	return id, nil
}

func (this *TrackerService) CheckSign(message proto.Message) (err error) {
	id, err := this.GetCryptoInfo(message)
	if err != nil {
		return err
	}
	log.Debugf("tk check sign pubKey: %v, rawData: %v, sign: %v", id.PubKey, id.RawData, id.Sign)
	return chainsdk.Verify(id.PubKey, id.RawData, id.Sign)
}

func (this *TrackerService) ReceiveAnnounceMessage(message proto.Message, from string) {
	log.Debug("receive announce message: ", reflect.TypeOf(message).String(), " from: ", from)
	id, err := this.GetCryptoInfo(message)
	log.Debugf("crypto info: %v, err: %v", id, err)
	if err != nil {
		return
	}
	log.Debugf("tk check sign pubKey: %v, rawData: %v, sign: %v", id.PubKey, id.RawData, id.Sign)
	if err := chainsdk.Verify(id.PubKey, id.RawData, id.Sign); err != nil {
		log.Errorf("check sign failed: %v", err)
		return
	}
	switch msg := message.(type) {
	case *tkpm.AnnounceRequestMessage:
		log.Debugf("msg type %s", msg.GetRequest().Event.String())
		annResp, err := this.onAnnounce(msg.GetRequest(), id)
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
			log.Debugf("HandleAnnounceResponseEvent %v %s", annResp, from)
			this.HandleAnnounceResponseEvent(annResp, from)
		}
	case *tkpm.AnnounceResponseMessage:
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
		this.ReceiveAnnounceResponseMessage(msg)
	}
}

func (this *TrackerService) ReceiveAnnounceResponseMessage(annRespMsg *tkpm.AnnounceResponseMessage) {
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

func (this *TrackerService) onAnnounce(aReq *tkpm.AnnounceRequest, id CryptoIdentity) (*tkpm.AnnounceResponse, error) {
	switch aReq.Event {
	case tkpm.AnnounceEvent_COMPLETE_TORRENT:
		return this.onAnnounceCompleteTorrent(aReq)
	case tkpm.AnnounceEvent_QUERY_TORRENT_PEERS:
		return this.onAnnounceQueryTorrentPeers(aReq, id)
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
	log.Debugf("put endpoint %v %v %v", wallet.ToBase58(), peer.IP, int(req.Port))
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
	log.Debugf("wallet %v %s", wallet, wallet.ToBase58())
	peer, err := storage.EDB.GetEndpoint(wallet.ToBase58())
	if err != nil {
		return nil, err
	}
	if peer == nil {
		return nil, errors.New("endpoint not registed")
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

func (this *TrackerService) onAnnounceQueryTorrentPeers(aReq *tkpm.AnnounceRequest, id CryptoIdentity) (*tkpm.AnnounceResponse, error) {
	req := aReq.GetTorrentPeersReq
	if req == nil {
		return nil, errors.New("request.GetTorrentPeersReq is nil")
	}
	if storage.TDB == nil {
		return nil, errors.New("storage.TDB is nil")
	}
	pass, _ := this.CheckWhiteListFn(string(req.InfoHash), types.AddressFromPubKey(id.PubKey))
	if !pass {
		return nil, fmt.Errorf("ban this query out whitelist")
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
		pi.Timestamp = time.Now()
		pi.Print()
		torrent.Peers[pi.NodeAddr.String()] = pi
		storage.TDB.PutTorrent(req.InfoHash, torrent)
	}

	piBinarys := pi.Serialize()
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
