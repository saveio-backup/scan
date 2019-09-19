package tracker

import (
	"sync"
	"fmt"
	"time"
	"errors"
	"reflect"

	tkpm "github.com/saveio/scan/p2p/messages/pb"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/scan/common"
	"github.com/anacrolix/dht/krpc"
	"github.com/gogo/protobuf/proto"
	"github.com/saveio/themis/common/log"
	tkActClient "github.com/saveio/scan/p2p/actor/tracker/client"
)

type TrackerService struct {
	AnnounceMessageMap *sync.Map
}

type AnnounceAction int32
type AnnounceMessageStatus int32

const (
	AnnounceActionRequest AnnounceAction = iota
	AnnounceActionResponse
)

const (
	AnnounceMessageRequestProcessed AnnounceMessageStatus = iota
	AnnounceMessageResponseDelivered
)

type AnnounceMessageItem struct {
	MsgID *tkpm.MessageID
	AnnRequest *tkpm.AnnounceRequest
	AnnResponse *tkpm.AnnounceResponse
	AnnResponseDelivered *tkpm.AnnounceResponseDelivered
	Action AnnounceAction // "request" or "response"
	ChStatus chan AnnounceMessageStatus // "request processed"、"response delivered"
}

type peerInfo struct {
	ID       common.PeerID
	Complete bool
	IP       [4]byte
	Port     uint16
	NodeAddr krpc.NodeAddr
}

type torrent struct {
	Leechers int32
	Seeders  int32
	Peers    map[string]*peerInfo
}

const (
	MAX_MSG_RETRY_TIMES = 4
	MAX_MSG_WAIT_SECONDS = 5
)

func NewTrackerService() *TrackerService {
	return &TrackerService{
		AnnounceMessageMap: new(sync.Map),
	}
}

func (this *TrackerService) Start() error {
	return nil
}

func (this *TrackerService) HandleAnnounceRequestEvent(arq *tkpm.AnnounceRequest, from string) (*tkpm.AnnounceResponse, error) {
	this.Sign()

	msgId := &tkpm.MessageID{
		MessageId: uint64(GetMsgID()),
	}

	annMsgItem := AnnounceMessageItem{
		MsgID: msgId,
		AnnRequest: arq,
		AnnResponse: nil,
		AnnResponseDelivered: nil,
		Action: AnnounceActionRequest,
		ChStatus: make(chan AnnounceMessageStatus),
	}
	this.AnnounceMessageMap.Store(msgId, annMsgItem)
	this.SignAndSend(from, msgId, annMsgItem.AnnRequest)

	var result *tkpm.AnnounceResponse
	var interval time.Duration = MAX_MSG_RETRY_TIMES
	t := time.NewTimer(interval * time.Second)

	// Waiting & Retry
	for {
		select {
		case <-time.After(MAX_MSG_WAIT_SECONDS * time.Second):
			log.Debugf("")
			msgItem, ok := this.AnnounceMessageMap.Load(msgId)
			if !ok {

			}
			if msgItem != nil {
				this.AnnounceMessageMap.Delete(msgId)
			}
			break
		case <-t.C:
			log.Debugf("[QueueSend] <-t.C Time: %s queue: %+v\n", time.Now().String(), annMsgItem.AnnRequest)
			this.SignAndSend(from, msgId, annMsgItem.AnnRequest)
			log.Warnf("Timeout retry for msg = %+v\n", annMsgItem.AnnRequest)
			t.Reset(interval * time.Second)
		case status := <-annMsgItem.ChStatus:
			msgItemInterface, ok := this.AnnounceMessageMap.Load(msgId)
			msgItem := msgItemInterface.(*AnnounceMessageItem)
			if msgItem.Action == AnnounceActionRequest && status == AnnounceMessageRequestProcessed {
				result = msgItem.AnnResponse
				if result != nil {
					go this.SignAndSend(from,  msgId, &tkpm.AnnounceResponseDelivered{
						MsgId: msgItem.MsgID,
						Signature: this.Sign(),
					})
				}
				this.AnnounceMessageMap.Delete(msgItem.MsgID)
				break
			}
		}
	}

	return result, nil
}

func (this * TrackerService) SignAndSend(from string, msgId *tkpm.MessageID, message proto.Message) error {
	signature, err := this.Sign()
	if err != nil {
		return errors.New("make message signature failed.")
	}

	switch message.(type) {
	case *tkpm.AnnounceRequest:
		tkActClient.P2pSend(from, &tkpm.AnnounceRequestMessage{
			MessageIdentifier: msgId,
			Request: message.(*tkpm.AnnounceRequest),
			Signature: signature,
		})
	case *tkpm.AnnounceResponse:
		tkActClient.P2pSend(from, &tkpm.AnnounceResponseMessage{
			MessageIdentifier: msgId,
			Response: message.(*tkpm.AnnounceResponse),
			Signature: signature,
		})
	case *tkpm.AnnounceResponseDelivered:
		tkActClient.P2pSend(from, &tkpm.AnnounceResponseDelivered{
			MsgId: msgId,
			Signature: signature,
		})
	default:
		return fmt.Errorf("Unknown message type to send")
	}

	return nil
}

func (this *TrackerService) Sign() error {
	return nil
}

func (this *TrackerService) CheckSign() error {
	return nil
}

func (this *TrackerService) ReceiveAnnounceMessage(message proto.Message, from string) {
	log.Debug("[NetComponent] Receive: ", reflect.TypeOf(message).String(), " From: ", from)
	switch message.(type) {
	case *tkpm.AnnounceRequestMessage:
		this.ReceiveAnnounceRequestMessage(message.(*tkpm.AnnounceRequestMessage), from)
	case *tkpm.AnnounceResponseMessage:
		this.ReceiveAnnounceResponseMessage(message.(*tkpm.AnnounceResponseMessage), from)
	case *tkpm.AnnounceResponseDelivered:
		this.ReceiveAnnounceResponseDeliveredMessage(message.(*tkpm.AnnounceResponseDelivered), from)
	}
}

func (this *TrackerService) ReceiveAnnounceResponseMessage(annRespMsg *tkpm.AnnounceResponseMessage, from string) {
	msgId := annRespMsg.MessageIdentifier
	msg, ok := this.AnnounceMessageMap.Load(msgId)
	if ok && msg != nil {
		msgItem := msg.(*AnnounceMessageItem)
		msgItem.AnnResponse = annRespMsg.Response
		this.AnnounceMessageMap.Store(msgId, msgItem)
		msgItem.ChStatus <- AnnounceMessageRequestProcessed
	}
}

func (this *TrackerService) ReceiveAnnounceResponseDeliveredMessage(annRespDlvMsg *tkpm.AnnounceResponseDelivered, from string) {
	msgId := annRespDlvMsg.MsgId
	msg, ok := this.AnnounceMessageMap.Load(msgId)
	if ok && msg != nil {
		msgItem := msg.(*AnnounceMessageItem)
		msgItem.AnnResponseDelivered = annRespDlvMsg
		this.AnnounceMessageMap.Store(msgId, msgItem)
		msgItem.ChStatus <- AnnounceMessageResponseDelivered
	}
}

func (this *TrackerService) ReceiveAnnounceRequestMessage(annReqMsg *tkpm.AnnounceRequestMessage, from string) {
	ar := annReqMsg.Request
	nodeAddr := krpc.NodeAddr{
		IP:   ar.IPAddress[:],
		Port: int(ar.Port),
	}
	t := s.getTorrent(ar.InfoHash)
	var pi *peerInfo
	if t != nil {
		pi = t.Peers[nodeAddr.String()]
	}

	switch req.Event {
	case tkpm.AnnounceRequest_EMPTY:
		this.onAnnounceStartedEmpty()
	case tkpm.AnnounceRequest_STARTED:
		this.onAnnounceStarted()
	case tkpm.AnnounceRequest_UPDATED:
		this.onAnnounceUpdated()
	case tkpm.AnnounceRequest_STOPED:
		this.onAnnounceStoped()
	case tkpm.AnnounceRequest_COMPLETED:
		this.onAnnounceCompleted()
	}

	// update
	t = s.getTorrent(ar.InfoHash)
	if t == nil || t.Peers == nil {
		err = s.respond(addr, ResponseHeader{
			TransactionId: h.TransactionId,
			Action:        ActionAnnounce,
		}, AnnounceResponseHeader{
			Interval: 900,
			Leechers: 0,
			Seeders:  0,
		}, []byte{})
		return
	}
	bm := func() encoding.BinaryMarshaler {
		ip := missinggo.AddrIP(addr)
		pNodeAddrs := make([]krpc.NodeAddr, 0)
		for _, pi := range t.Peers {
			if !pi.Complete {
				continue
			}
			nodeAddIp := pi.NodeAddr.IP.To4()
			if nodeAddIp == nil {
				nodeAddIp = pi.NodeAddr.IP.To16()
			}
			pNodeAddrs = append(pNodeAddrs, krpc.NodeAddr{
				IP:   nodeAddIp,
				Port: pi.NodeAddr.Port,
			})
			if ar.NumWant != -1 && len(pNodeAddrs) >= int(ar.NumWant) {
				break
			}
		}
		if ip.To4() != nil {
			return krpc.CompactIPv4NodeAddrs(pNodeAddrs)
		}
		return krpc.CompactIPv6NodeAddrs(pNodeAddrs)
	}()
	b, err = bm.MarshalBinary()
	if err != nil {
		panic(err)
	}
	err = s.respond(addr, ResponseHeader{
		TransactionId: h.TransactionId,
		Action:        ActionAnnounce,
	}, AnnounceResponseHeader{
		Interval: 900,
		Leechers: t.Leechers,
		Seeders:  t.Seeders,
	}, b)

	annResponse := &tkpm.AnnounceResponse{
		Interval: 900,
		Leechers: t.Leechers,
		Seeders: t.Seeders,

	}

	msgId := annReqMsg.MessageIdentifier
	annMsgItem := AnnounceMessageItem{
		MsgId: msgId,
		AnnRequest: nil,
		AnnResponse: nil,
		AnnResponseDelivered: nil,
		Action: AnnounceActionResponse,
		ChStatus: make(chan AnnounceMessageStatus),
	}
	this.AnnounceMessageMap.LoadAndStore(msgId, AnnMsgItem)
	this.SignAndSend(from, msgId, AnnMsgItem.AnnResponse)

	var interval time.Duration = MAX_MSG_RETRY_TIMES
	ti := time.NewTimer(interval * time.Second)
	// Waiting & Retry
	for {
		select {
		case <-time.After(MAX_MSG_WAIT_SECONDS * time.Second):
			log.Debugf("")
			msgItem := this.AccounceMessageMap.Load(msgID)
			if msgItem != nil {
				this.AnnounceMessageMap.Delete(msgItem.MsgId)
			}
			break
		case <-ti.C:
			log.Debugf("[QueueSend] <-t.C Time: %s queue: %p\n", time.Now().String(), queue)
			this.SendAsync(&AnnMsgItem)
			log.Warnf("Timeout retry for msg = %+v\n", msg)
			ti.Reset(interval * time.Second)
		case status := <-this.AnnounceMessageMap.Load(msgId).ChStatus:
			msgItem := this.AccounceMessageMap.Load(msgID)
			if msgItem.Action == AnnounceActionResponse && status == AnnounceMessageResponseDelivered {
				if msgItem.AnnResponseDelivered != nil && msgItem.AnnResponseMsg.Response != nil && msgItem.AnnResponseDelivered.MsgId == msgItem.msgId {
					this.AnnounceMessageMap.Delete(msgItem.MsgId)
				}
				break
			}
		}
	}
}



func (this *TrackerService) onAnnounceStarted(req *tkpm.AnnounceRequest, pi *peerInfo) {
	if pi != nil {
		s.onAnnounceUpdated(ar, pi)
		return
	}

	//ip := make(net.IP, 4)
	//binary.BigEndian.PutUint32(ip, ar.IPAddress)
	peer := krpc.NodeAddr{
		IP:   ar.IPAddress[:],
		Port: int(ar.Port),
	}

	t := s.getTorrent(ar.InfoHash)
	if t == nil {
		t = &torrent{}
	}

	if ar.Left == 0 {
		t.Seeders++
	} else {
		t.Leechers++
	}
	if t.Peers == nil {
		t.Peers = make(map[string]*peerInfo, 0)
	}

	pi = &peerInfo{
		ID:       ar.PeerId,
		Complete: ar.Left == 0,
		IP:       ar.IPAddress,
		Port:     ar.Port,
		NodeAddr: peer,
	}

	t.Peers[peer.String()] = pi
	bt, err := json.Marshal(t)
	if err != nil {
		log.Fatalf("json Marshal error:%s", err)
	}
	storage.TDB.Put(ar.InfoHash[:], bt)
	m := &pm.Torrent{
		InfoHash: ar.InfoHash[:],
		Torrent:  bt,
		Type:     0,
	}
	s.p2p.Tell(m)
}

func (this *TrackerService) onAnnounceUpdated(req *tkpm.AnnounceRequest, pi *peerInfo) {
	if pi == nil {
		s.onAnnounceStarted(ar, nil)
		return
	}
	t := s.getTorrent(ar.InfoHash)
	if t == nil {
		return
	}
	if !pi.Complete && ar.Left == 0 {
		t.Leechers -= 1
		t.Seeders += 1
		pi.Complete = true
	}
	t.Peers[pi.NodeAddr.String()] = pi
	bt, err := json.Marshal(t)
	if err != nil {
		log.Fatalf("json Marshal error:%s", err)
	}
	storage.TDB.Put(ar.InfoHash[:], bt)
	m := &pm.Torrent{
		InfoHash: ar.InfoHash[:],
		Torrent:  bt,
		Type:     0,
	}
	s.p2p.Tell(m)
}

func (this *TrackerService) onAnnounceStoped(req *tkpm.AnnounceRequest, pi *peerInfo) {
	if pi == nil {
		return
	}
	t := s.getTorrent(ar.InfoHash)
	if t == nil {
		return
	}
	if pi.Complete {
		t.Seeders -= 1
	} else {
		t.Leechers -= 1
	}
	delete(t.Peers, pi.NodeAddr.String())
	bt, err := json.Marshal(t)
	if err != nil {
		log.Fatalf("json Marshal error:%s", err)
	}
	storage.TDB.Put(ar.InfoHash[:], bt)
	m := &pm.Torrent{
		InfoHash: ar.InfoHash[:],
		Torrent:  bt,
		Type:     0,
	}
	s.p2p.Tell(m)
	//messageBus.MsgBus.TNTBox <- &messageBus.TorrenMsg{
	//	InfoHash:     ar.InfoHash[:],
	//	BytesTorrent: bt,
	//}
}

func (this *TrackerService) onAnnounceCompleted(req *tkpm.AnnounceRequest, pi *peerInfo) {
	if pi == nil {
		s.onAnnounceStarted(ar, nil)
		return
	}
	if pi.Complete {
		s.onAnnounceUpdated(ar, pi)
		return
	}
	t := s.getTorrent(ar.InfoHash)
	t.Seeders += 1
	t.Leechers -= 1
	pi.Complete = true
	t.Peers[pi.NodeAddr.String()] = pi
	bt, err := json.Marshal(t)
	if err != nil {
		log.Fatalf("json Marshal error:%s", err)
	}
	storage.TDB.Put(ar.InfoHash[:], bt)
	m := &pm.Torrent{
		InfoHash: ar.InfoHash[:],
		Torrent:  bt,
		Type:     0,
	}
	s.p2p.Tell(m)
	//messageBus.MsgBus.TNTBox <- &messageBus.TorrenMsg{
	//	InfoHash:     ar.InfoHash[:],
	//	BytesTorrent: bt,
	//}
}

func (this *TrackerService) getTorrent(infoHash common.MetaInfoHash) *torrent {
	var t torrent
	v, err := storage.TDB.Get(infoHash[:])
	json.Unmarshal(v, &t)
	if v == nil || err != nil {
		return nil
	} else {
		return &t
	}
}

func marshal(parts ...interface{}) (ret []byte, err error) {
	var buf bytes.Buffer
	for _, p := range parts {
		err = binary.Write(&buf, binary.BigEndian, p)
		if err != nil {
			return
		}
	}
	ret = buf.Bytes()
	return
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