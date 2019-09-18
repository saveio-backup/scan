package tracker

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"

	"github.com/saveio/themis/crypto/keypair"

	"github.com/anacrolix/dht/krpc"
	"github.com/anacrolix/missinggo"
	"github.com/ontio/ontology-eventbus/actor"
	pm "github.com/saveio/scan/p2p/actor/messages"
	"github.com/saveio/scan/storage"
	tkComm "github.com/saveio/scan/tracker/common"
	chainsdk "github.com/saveio/themis-go-sdk/utils"
	theComm "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

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
		log.Error("[return ReadFrom]: ", err.Error())
		return errors.New("return ReadFrom")
	}
	r := bytes.NewReader(b[:n])
	var h RequestHeader
	err = readBody(r, &h)
	if err != nil {
		log.Error("[return readBody]: ", err.Error())
		return errors.New("return readBody")
	}

	switch h.Action {
	case ActionConnect:
		log.Debugf("tracker.server.ActionConnect")
		if h.ConnectionId != connectRequestConnectionId {
			return errors.New("return connectRequestConnectionId")
		}
		connId := s.newConn()
		err = s.respond(addr, ResponseHeader{
			Action:        ActionConnect,
			TransactionId: h.TransactionId,
		}, ConnectionResponse{
			ConnectionId: connId,
		})
		if err != nil {
			log.Debugf("tracker.server.return ActionConnect respond return err: %v\n", err)
		}
		return
	case ActionAnnounce:
		log.Debugf("tracker.server. Accepted ActionAnnounce")
		if _, ok := s.conns[h.ConnectionId]; !ok {
			err := s.respond(addr, ResponseHeader{
				TransactionId: h.TransactionId,
				Action:        ActionError,
			}, []byte("not connected"))
			if err != nil {
				log.Debugf("tracker.server.return ActionAnnounce respond 1 not connected, err: %v\n", err)
			}
			return err
		}
		var ar AnnounceRequest
		err = readBody(r, &ar)
		if err != nil {
			log.Debugf("tracker.server.return readBody, err: %v\n", err)
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("read announcerequest body failed"))
		}
		log.Debugf("tracker.server.ActionAnnounce: %v, wallet: %s, fileHash: %s ", ar, ar.Wallet.ToBase58(), string(ar.InfoHash[:]))

		if len(ar.InfoHash) == 0 {
			log.Debugf("tracker.server.return len(InfoHash) == 0, err: %v\n", err)
			return
		}

		pubKey, sigData, err := ParseOptions(r)
		log.Debugf("ParseOptions: pubKey %v, sigData %v, err %v\n", pubKey, sigData, err)
		if err != nil {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("parse announcerequest options failed"))
		}

		nodeAddr := krpc.NodeAddr{IP: ar.IPAddress[:], Port: int(ar.Port)}

		var rawData []byte
		if ar.Event.String() == "completed" {
			rawData, err = json.Marshal(ActionTorrentCompleteParams{InfoHash: ar.InfoHash, IP: nodeAddr.IP, Port: ar.Port})
		} else {
			rawData, err = json.Marshal(ActionGetTorrentPeersParams{InfoHash: ar.InfoHash, NumWant: ar.NumWant, Left: ar.Left})
		}
		if err != nil {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("json.Marshal rawData failed"))
		}

		err = chainsdk.Verify(pubKey, rawData, sigData)
		if err != nil {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("verify signature failed"))
		}

		pi, err := storage.TDB.GetTorrentPeerByFileHashAndNodeAddr(ar.InfoHash[:], nodeAddr.String())
		log.Debugf("pi %v, err %v\n", pi, err)
		if err != nil {
			log.Debugf("tracker.server.ActionAnnounce pi: %v, nodeAddr: %s", pi, nodeAddr.String())
		}

		var announceErr error
		switch ar.Event.String() {
		case "started":
			announceErr = s.onAnnounceStarted(&ar, pi)
		case "updated":
			announceErr = s.onAnnounceUpdated(&ar, pi)
		case "stopped":
			announceErr = s.onAnnounceStopped(&ar, pi)
		case "completed":
			announceErr = s.onAnnounceCompleted(&ar, pi)
		}
		if announceErr != nil {
			log.Debugf("tracker.server.ActionAnnounce announceErr: %v\n", err)
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionAnnounce}, AnnounceResponseHeader{Interval: 800, Leechers: 0, Seeders: 0}, []byte{})
		}
		// update
		peers, leechers, seeders, err := storage.TDB.GetTorrentPeersByFileHash(ar.InfoHash[:], ar.NumWant)
		if err != nil {
			return s.respond(addr, ResponseHeader{
				TransactionId: h.TransactionId,
				Action:        ActionAnnounce,
			}, AnnounceResponseHeader{
				Interval: 900,
				Leechers: 0,
				Seeders:  0,
			}, []byte{})
		}
		bm := func() encoding.BinaryMarshaler {
			ip := missinggo.AddrIP(addr)
			pNodeAddrs := make([]krpc.NodeAddr, 0)
			for _, pi := range peers {
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
		log.Debugf("tracker.server.Accepted bm: %v\n", bm)
		b, err = bm.MarshalBinary()
		if err != nil {
			panic(err)
		}

		return s.respond(addr, ResponseHeader{
			TransactionId: h.TransactionId,
			Action:        ActionAnnounce,
		}, AnnounceResponseHeader{
			Interval: 900,
			Leechers: leechers,
			Seeders:  seeders,
		}, b)
	case ActionReg:
		log.Info("TrackerServer Accepted ActionReg")
		if _, ok := s.conns[h.ConnectionId]; !ok {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("not connected"))
		}
		var ar AnnounceRequest
		err = readBody(r, &ar)
		if err != nil {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("read announcerequest body failed"))
		}

		pubKey, sigData, err := ParseOptions(r)
		log.Debugf("ParseOptions: pubKey %v, sigData %v, err %v\n", pubKey, sigData, err)
		if err != nil {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("parse announcerequest options failed"))
		}

		nodeAddr := krpc.NodeAddr{IP: ar.IPAddress[:], Port: int(ar.Port)}
		rawData, err := json.Marshal(ActionEndpointRegParams{Wallet: ar.Wallet, IP: nodeAddr.IP, Port: ar.Port})
		if err != nil {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("json.Marshal rawData failed"))
		}

		err = chainsdk.Verify(pubKey, rawData, sigData)
		if err != nil {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("verify signature failed"))
		}

		if err != nil || ar.Wallet == theComm.ADDRESS_EMPTY {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("read announce request buffer failed"))
		}
		err = storage.EDB.PutEndpoint(ar.Wallet.ToBase58(), nodeAddr.IP, int(ar.Port))
		if err != nil {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("db put endpoint failed"))
		}
		s.p2p.Tell(&pm.Endpoint{WalletAddr: ar.Wallet.ToBase58(), HostPort: nodeAddr.String(), Type: 0})
		log.Infof("ActionReg wallAddr:%s, nodeAddr:%s", ar.Wallet.ToBase58(), nodeAddr.String())
		return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionReg}, AnnounceResponseHeader{IPAddress: ipAddr, Port: ar.Port, Wallet: ar.Wallet})
	case ActionUnReg:
		log.Info("TrackerServer Accepted ActionUnReg")
		if _, ok := s.conns[h.ConnectionId]; !ok {
			return s.respond(addr, ResponseHeader{TransactionId: h.TransactionId, Action: ActionError}, []byte("not connected"))
		}
		var ar AnnounceRequest
		err = readBody(r, &ar)
		if err != nil || ar.Wallet == theComm.ADDRESS_EMPTY {
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
		if ar.Wallet == theComm.ADDRESS_EMPTY || err != nil || nodeAddr == nil {
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
		if ar.Wallet == theComm.ADDRESS_EMPTY {
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
		log.Infof("Tracker client  reg success,wallet:%s,nodeAddr:%s,nodeType:%d", theComm.ToHexString(ar.Wallet[:]), nodeAddr.String(), ar.NodeType)
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

func (s *Server) onAnnounceStarted(ar *AnnounceRequest, pi *storage.PeerInfo) error {
	if pi != nil {
		return s.onAnnounceUpdated(ar, pi)
	}

	nodeAddr := krpc.NodeAddr{IP: ar.IPAddress[:], Port: int(ar.Port)}
	pi = &storage.PeerInfo{
		ID:       storage.PeerID(ar.PeerId),
		Complete: ar.Left == 0,
		IP:       ar.IPAddress,
		Port:     ar.Port,
		NodeAddr: nodeAddr,
	}
	err := storage.TDB.AddTorrentPeer(ar.InfoHash[:], ar.Left, nodeAddr.String(), pi)
	if err != nil {
		return err
	}

	piBinarys := pi.Serialize()
	pi.Print()
	log.Debugf("PeerInfo Binary: %v\n", piBinarys)
	s.p2p.Tell(&pm.Torrent{InfoHash: ar.InfoHash[:], Left: ar.Left, Peerinfo: piBinarys, Type: 0})
	return err
}

func (s *Server) onAnnounceUpdated(ar *AnnounceRequest, pi *storage.PeerInfo) error {
	if pi == nil {
		return s.onAnnounceStarted(ar, nil)
	}

	err := storage.TDB.AddTorrentPeer(ar.InfoHash[:], ar.Left, pi.NodeAddr.String(), pi)
	if err != nil {
		return err
	}

	piBinarys := pi.Serialize()
	pi.Print()
	log.Debugf("PeerInfo Binary: %v\n", piBinarys)
	s.p2p.Tell(&pm.Torrent{InfoHash: ar.InfoHash[:], Left: ar.Left, Peerinfo: piBinarys, Type: 0})
	return err
}

func (s *Server) onAnnounceStopped(ar *AnnounceRequest, pi *storage.PeerInfo) error {
	if pi == nil {
		return nil
	}

	err := storage.TDB.DelTorrentPeer(ar.InfoHash[:], pi)
	if err != nil {
		return err
	}

	piBinarys := pi.Serialize()
	pi.Print()
	log.Debugf("PeerInfo Binary: %v\n", piBinarys)
	s.p2p.Tell(&pm.Torrent{InfoHash: ar.InfoHash[:], Left: ar.Left, Peerinfo: piBinarys, Type: 0})
	return err
}

func (s *Server) onAnnounceCompleted(ar *AnnounceRequest, pi *storage.PeerInfo) error {
	if pi == nil {
		return s.onAnnounceStarted(ar, nil)
	}
	if pi.Complete {
		return s.onAnnounceUpdated(ar, pi)
	}
	t, err := storage.TDB.GetTorrent(ar.InfoHash[:])
	if err != nil {
		return err
	}
	t.Seeders += 1
	t.Leechers -= 1
	pi.Complete = true
	t.Peers[pi.NodeAddr.String()] = pi
	storage.TDB.PutTorrent(ar.InfoHash[:], t)

	piBinarys := pi.Serialize()
	pi.Print()
	log.Debugf("PeerInfo Binary: %v\n", piBinarys)
	s.p2p.Tell(&pm.Torrent{InfoHash: ar.InfoHash[:], Left: ar.Left, Peerinfo: piBinarys, Type: 0})
	return err
}

func ParseOptions(r io.Reader) (pubKey keypair.PublicKey, signData []byte, err error) {
	i := 0
	for i < 2 {
		tlv, err := tkComm.Read(r)
		if err != nil {
			break
		}

		switch tlv.Tag {
		case tkComm.TLV_TAG_PUBLIC_KEY:
			pubKey, err = keypair.DeserializePublicKey(tlv.GetValue())
			if err != nil {
				break
			}
		case tkComm.TLV_TAG_SIGNATURE:
			signData = tlv.GetValue()
		}
		i++
	}
	return
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
