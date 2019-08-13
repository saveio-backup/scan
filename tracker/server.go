package tracker

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"

	"encoding/json"

	"github.com/anacrolix/dht/krpc"
	"github.com/anacrolix/missinggo"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/scan/common"
	pm "github.com/saveio/scan/messages/protoMessages"
	"github.com/saveio/scan/storage"
	Ccomon "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

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

type Server struct {
	pc    net.PacketConn
	conns map[int64]struct{}
	p2p   *actor.PID
}

// NewServer
func NewServer() *Server {
	return &Server{
		conns: make(map[int64]struct{}, 0),
	}
}

// SetPacketConn
func (s *Server) SetPacketConn(p net.PacketConn) {
	s.pc = p
}

// Run tracker server
func (s *Server) Run() error {
	log.Info("Tracker is accepting request service in goroutine")
	for {
		err := s.Accepted()
		if err != nil {
			log.Errorf("tracker accepted err:%s\n", err)
		}
	}
}

// Accpted request service
func (s *Server) Accepted() (err error) {
	b := make([]byte, 0x10000)
	n, addr, err := s.pc.ReadFrom(b)
	if err != nil {
		log.Error("[ReadFrom]: ", err.Error())
		return
	}
	r := bytes.NewReader(b[:n])
	var h RequestHeader
	err = readBody(r, &h)
	if err != nil {
		log.Error("[readBody]: ", err.Error())
		return
	}

	switch h.Action {
	case ActionConnect:
		log.Info("TrackerServer Accepted ActionConnect")
		if h.ConnectionId != connectRequestConnectionId {
			return
		}
		connId := s.newConn()
		err = s.respond(addr, ResponseHeader{
			Action:        ActionConnect,
			TransactionId: h.TransactionId,
		}, ConnectionResponse{
			ConnectionId: connId,
		})
		return
	case ActionAnnounce:
		log.Info("TrackerServer Accepted ActionAnnounce")
		if _, ok := s.conns[h.ConnectionId]; !ok {
			s.respond(addr, ResponseHeader{
				TransactionId: h.TransactionId,
				Action:        ActionError,
			}, []byte("not connected"))
			return
		}
		var ar AnnounceRequest
		err = readBody(r, &ar)
		if err != nil {
			return
		}

		if len(ar.InfoHash) == 0 {
			err = fmt.Errorf("no info hash")
			return
		}
		//ip := make(net.IP, 4)
		//binary.BigEndian.PutUint32(ip, ar.IPAddress)
		nodeAddr := krpc.NodeAddr{
			IP:   ar.IPAddress[:],
			Port: int(ar.Port),
		}
		t := s.getTorrent(ar.InfoHash)
		var pi *peerInfo
		if t != nil {
			log.Infof("tracker.server.ActionAnnounce s.getTorrent: %v", t.Peers)
			pi = t.Peers[nodeAddr.String()]
		}
		log.Info("tracker.server.ActionAnnounce pi: %v, nodeAddr: %s", pi, nodeAddr.String())
		switch ar.Event.String() {
		case "started":
			s.onAnnounceStarted(&ar, pi)
		case "updated":
			s.onAnnounceUpdated(&ar, pi)
		case "stopped":
			s.onAnnounceStopped(&ar, pi)
		case "completed":
			s.onAnnounceCompleted(&ar, pi)
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
		log.Infof("tracker.server.Accepted bm: %v\n", bm)
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
		return
	case ActionReg:
		log.Info("TrackerServer Accepted ActionReg")
		if _, ok := s.conns[h.ConnectionId]; !ok {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("not connected"))
		}
		var ar AnnounceRequest
		err = readBody(r, &ar)
		if err != nil || ar.Wallet == Ccomon.ADDRESS_EMPTY {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("read announce request buffer failed"))
		}
		nodeAddr := krpc.NodeAddr{IP: ar.IPAddress[:], Port: int(ar.Port)}
		err = storage.EDB.PutEndpoint(ar.Wallet.ToBase58(), nodeAddr.IP, int(ar.Port))
		if err != nil {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("db put endpoint failed"))
		}
		s.p2p.Tell(&pm.Registry{WalletAddr: ar.Wallet.ToBase58(), HostPort: nodeAddr.String(), Type: 0})
		log.Infof("ActionReg wallAddr:%s, nodeAddr:%s", ar.Wallet.ToBase58(), nodeAddr.String())
		return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionReg}, AnnounceResponseHeader{IPAddress: ipAddr, Port: ar.Port, Wallet: ar.Wallet})
	case ActionUnReg:
		log.Info("TrackerServer Accepted ActionUnReg")
		if _, ok := s.conns[h.ConnectionId]; !ok {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("not connected"))
		}
		var ar AnnounceRequest
		err = readBody(r, &ar)
		if err != nil || ar.Wallet == Ccomon.ADDRESS_EMPTY {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("read announce request buffer failed"))
		}

		exist, err := storage.EDB.GetEndpoint(ar.Wallet.ToBase58())
		if exist == nil || err != nil {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("not exists"))
		}

		err = storage.EDB.DelEndpoint(ar.Wallet.ToBase58())
		if err != nil {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("db del endpoint failed"))
		}
		s.p2p.Tell(&pm.UnRegistry{WalletAddr: ar.Wallet.ToBase58(), Type: 0})
		return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionUnReg}, AnnounceResponseHeader{Wallet: ar.Wallet})
	case ActionReq:
		log.Info("TrackerServer Accepted ActionReq")
		if _, ok := s.conns[h.ConnectionId]; !ok {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("not connected"))
		}
		var ar AnnounceRequest
		err = readBody(r, &ar)
		if err != nil {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("read announce request buffer failed"))
		}
		nodeAddr, err := storage.EDB.GetEndpoint(ar.Wallet.ToBase58())
		if ar.Wallet == Ccomon.ADDRESS_EMPTY || err != nil || nodeAddr == nil {
			// return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionReq}, AnnounceResponseHeader{Wallet: ar.Wallet})
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("endpoint not found"))
		}
		ipAddr := ipconvert(nodeAddr.NodeAddr.IP.To4())
		log.Infof("ActionReq wallAddr: %s, nodeAddr: %s", ar.Wallet.ToBase58(), nodeAddr.NodeAddr.String())
		return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionReq}, AnnounceResponseHeader{Wallet: ar.Wallet, IPAddress: ipAddr, Port: uint16(nodeAddr.NodeAddr.Port)})
	case ActionRegNodeType:
		log.Info("TrackerServer Accepted ActionRegNodeType")
		if _, ok := s.conns[h.ConnectionId]; !ok {
			s.respond(addr, ResponseHeader{
				TransactionId: h.TransactionId,
				Action:        ActionError,
			}, []byte("not connected"))
			return
		}
		var ar AnnounceRequest
		err = readBody(r, &ar)
		if err != nil {
			return
		}
		if ar.Wallet == Ccomon.ADDRESS_EMPTY {
			err = fmt.Errorf("nil walletAddr")
			return
		}
		nodeAddr := krpc.NodeAddr{
			IP:   ar.IPAddress[:],
			Port: int(ar.Port),
		}

		err = SaveNodeInfo(ar.NodeType, ar.Wallet, nodeAddr.String())
		if err != nil {
			log.Errorf("SaveNodeInfo error: %s", err.Error())
			return
		}

		ipAddr := ipconvert(nodeAddr.IP)
		err = s.respond(addr, ResponseHeader{
			TransactionId: h.TransactionId,
			Action:        ActionReg,
		}, AnnounceResponseHeader{
			IPAddress: ipAddr,
			Port:      ar.Port,
			Wallet:    ar.Wallet,
		})
		log.Infof("Tracker client  reg success,wallet:%s,nodeAddr:%s,nodeType:%d", Ccomon.ToHexString(ar.Wallet[:]), nodeAddr.String(), ar.NodeType)
		return err

	case ActionGetNodesByType:
		log.Info("TrackerServer Accepted ActionGetNodesByType")
		if _, ok := s.conns[h.ConnectionId]; !ok {
			s.respond(addr, ResponseHeader{
				TransactionId: h.TransactionId,
				Action:        ActionError,
			}, []byte("not connected"))
			return
		}
		var ar AnnounceRequest
		err = readBody(r, &ar)
		if err != nil {
			return
		}

		var nodesInfo []byte
		nodesInfo, err = GetNodesInfo(ar.NodeType)
		if err != nil {
			err = fmt.Errorf("SaveNodeInfo error: %s", err.Error())
			return
		}

		err = s.respond(addr, ResponseHeader{
			TransactionId: h.TransactionId,
			Action:        ActionReg,
		}, AnnounceResponseHeader{}, nodesInfo)
		log.Debugf("Server ActionGetNodesByType: %v", nodesInfo)
		log.Info("Tracker client lookUp success")
		return err
	default:
		log.Info("TrackerServer Accepted default")
		err = fmt.Errorf("unhandled action: %d", h.Action)
		s.respond(addr, ResponseHeader{
			TransactionId: h.TransactionId,
			Action:        ActionError,
		}, []byte("unhandled action"))
		return
	}
}

func (s *Server) onAnnounceStarted(ar *AnnounceRequest, pi *peerInfo) {
	if pi != nil {
		s.onAnnounceUpdated(ar, pi)
		return
	}

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
	log.Infof("tracker.server.onAnnounceStarted torrent.Peers: %v\n", t.Peers)
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

func (s *Server) onAnnounceUpdated(ar *AnnounceRequest, pi *peerInfo) {
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
	log.Infof("tracker.server.onAnnounceUpdated torrent.Peers: %v\n", t.Peers)
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

func (s *Server) onAnnounceStopped(ar *AnnounceRequest, pi *peerInfo) {
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

func (s *Server) onAnnounceCompleted(ar *AnnounceRequest, pi *peerInfo) {
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

func (s *Server) getTorrent(infoHash common.MetaInfoHash) *torrent {
	var t torrent
	v, err := storage.TDB.Get(infoHash[:])
	json.Unmarshal(v, &t)
	log.Infof("tracker.server.getTorrent: %v", t)
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

func (s *Server) respond(addr net.Addr, rh ResponseHeader, parts ...interface{}) (err error) {
	b, err := marshal(append([]interface{}{rh}, parts...)...)
	if err != nil {
		return
	}
	_, err = s.pc.WriteTo(b, addr)
	return
}

func (s *Server) newConn() (ret int64) {
	ret = rand.Int63()
	if s.conns == nil {
		s.conns = make(map[int64]struct{})
	}
	s.conns[ret] = struct{}{}
	return
}

// SetPID sets p2p actor
func (this *Server) SetPID(pid *actor.PID) {
	this.p2p = pid
}

// GetPID return msgBus actor
func (this *Server) GetPID() *actor.PID {
	return this.p2p
}
