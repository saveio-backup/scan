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
	Ccomon "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/scan/channel"
	"github.com/saveio/scan/common"
	"github.com/saveio/scan/common/config"
	pm "github.com/saveio/scan/messages/protoMessages"
	"github.com/saveio/scan/storage"
	"github.com/ontio/ontology-eventbus/actor"
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
	case ActionReg:
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
		//ip := make(net.IP, 4)
		//binary.BigEndian.PutUint32(ip, ar.IPAddress)
		nodeAddr := krpc.NodeAddr{
			IP:   ar.IPAddress[:],
			Port: int(ar.Port),
		}
		nb, err := nodeAddr.MarshalBinary()
		if err != nil {
			err = fmt.Errorf("nodeAddr marshal error")
			return err
		}
		err = storage.TDB.Put(ar.Wallet[:], nb)
		m := &pm.Registry{
			WalletAddr: ar.Wallet.ToHexString(),
			HostPort:   nodeAddr.String(),
		}
		s.p2p.Tell(m)
		if err != nil {
			return err
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
		channel.GlbChannelSvr.Channel.SetHostAddr(m.WalletAddr, config.DefaultConfig.ChannelConfig.ChannelProtocol+"://"+m.HostPort)
		log.Infof("Tracker client  reg success,wallet:%s,nodeAddr:%s", Ccomon.ToHexString(ar.Wallet[:]), nodeAddr.String())
		return err
	case ActionUnReg:
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
		var exist bool
		exist, err = storage.TDB.Has(ar.Wallet[:])
		if !exist || err != nil {
			log.Errorf("This wallet:%s is not exist!", ar.Wallet)
			return
		}
		err = storage.TDB.Delete(ar.Wallet[:])
		m := &pm.UnRegistry{
			WalletAddr: ar.Wallet.ToHexString(),
		}
		s.p2p.Tell(m)
		s.respond(addr, ResponseHeader{
			TransactionId: h.TransactionId,
			Action:        ActionUnReg,
		}, AnnounceResponseHeader{
			Wallet: ar.Wallet,
		})
		log.Debugf("Tracker client  unReg success,wallet:%s", ar.Wallet)
		return

	case ActionReq:
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
		nb, err := storage.TDB.Get(ar.Wallet[:])
		if err != nil {
			return err
		}
		var nodeAddr krpc.NodeAddr
		nodeAddr.UnmarshalBinary(nb)
		ipAddr := ipconvert(nodeAddr.IP)
		err = s.respond(addr, ResponseHeader{
			TransactionId: h.TransactionId,
			Action:        ActionReq,
		}, AnnounceResponseHeader{
			Wallet:    ar.Wallet,
			IPAddress: ipAddr,
			Port:      uint16(nodeAddr.Port),
		})

		log.Debugf("Tracker client  req success,wallet:%s,nodeAddr:%s", ar.Wallet, string(nb))
		return err

	case ActionUpdate:
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
		//ip := make(net.IP, 4)
		//binary.BigEndian.PutUint32(ip, ar.IPAddress)
		nodeAddr := krpc.NodeAddr{
			IP:   ar.IPAddress[:],
			Port: int(ar.Port),
		}
		var exist bool
		exist, err = storage.TDB.Has(ar.Wallet[:])
		if !exist || err != nil {
			log.Errorf("This wallet:%s is not exist!", ar.Wallet)
			return
		}
		var nb []byte
		nb, err = nodeAddr.MarshalBinary()
		if err != nil {
			return err
		}
		err = storage.TDB.Put(ar.Wallet[:], nb)
		if err != nil {
			return err
		}
		m := &pm.Registry{
			WalletAddr: ar.Wallet.ToHexString(),
			HostPort:   nodeAddr.String(),
		}
		s.p2p.Tell(m)
		ipAddr := ipconvert(nodeAddr.IP)
		log.Debugf("Tracker client  update success,wallet:%s,nodeAddr:%s", ar.Wallet, nodeAddr.String())
		s.respond(addr, ResponseHeader{
			TransactionId: h.TransactionId,
			Action:        ActionUpdate,
		}, AnnounceResponseHeader{
			Wallet:    ar.Wallet,
			IPAddress: ipAddr,
			Port:      uint16(nodeAddr.Port),
		})
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
	bt, err := json.Marshal(t)
	if err != nil {
		log.Fatalf("json Marshal error:%s", err)
	}
	storage.TDB.Put(ar.InfoHash[:], bt)
	m := &pm.Torrent{
		InfoHash: ar.InfoHash[:],
		Torrent:  bt,
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
