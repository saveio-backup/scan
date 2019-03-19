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
	"github.com/oniio/oniChain/common/log"
	"github.com/oniio/oniDNS/common"
	"github.com/oniio/oniDNS/storage"
)

type peerInfo struct {
	ID       common.PeerID
	Complete bool
	IP       uint32
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
	ls    *storage.LevelDBStore
	Net   *Network
}

// NewServer
func NewServer(path string) *Server {
	nls, err := storage.NewLevelDBStore(path)
	if err != nil {
		log.Errorf("init torrent cache err:%s\n", err)
		return nil
	}
	return &Server{
		conns: make(map[int64]struct{}, 0),
		ls: nls,
	}
}

// SetPacketConn
func (s *Server) SetPacketConn(p net.PacketConn) {
	s.pc = p
}

// Run tracker server
func (s *Server) Run() error {
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
		return
	}
	r := bytes.NewReader(b[:n])
	var h RequestHeader
	err = readBody(r, &h)
	if err != nil {
		return
	}
	switch h.Action {
	case ActionConnect:
		if h.ConnectionId != connectRequestConnectionId {
			return
		}
		connId := s.newConn()
		err = s.respond(addr, ResponseHeader{
			ActionConnect,
			h.TransactionId,
		}, ConnectionResponse{
			connId,
		})
		return
	case ActionAnnounce:
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
		//
		if len(ar.InfoHash) == 0 {
			err = fmt.Errorf("no info hash")
			return
		}
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, ar.IPAddress)
		nodeAddr := krpc.NodeAddr{
			IP:   ip,
			Port: int(ar.Port),
		}
		t := s.getTorrent(ar.InfoHash)
		var pi *peerInfo
		if t != nil {
			pi = t.Peers[nodeAddr.String()]
		}
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
	default:
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

	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, ar.IPAddress)
	peer := krpc.NodeAddr{
		IP:   ip,
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
	s.ls.Put(ar.InfoHash[:], bt)
	s.Net.SyncTorrent(ar.InfoHash[:], bt)
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
	bt, err := json.Marshal(t)
	if err != nil {
		log.Fatalf("json Marshal error:%s", err)
	}
	s.ls.Put(ar.InfoHash[:], bt)
	s.Net.SyncTorrent(ar.InfoHash[:], bt)
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
	s.ls.Put(ar.InfoHash[:], bt)
	s.Net.SyncTorrent(ar.InfoHash[:], bt)
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
	s.ls.Put(ar.InfoHash[:], bt)
	s.Net.SyncTorrent(ar.InfoHash[:], bt)
}

func (s *Server) getTorrent(infoHash common.MetaInfoHash) *torrent {
	var t torrent
	v, err := s.ls.Get(infoHash[:])
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

