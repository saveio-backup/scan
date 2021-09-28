package tracker

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/anacrolix/dht/krpc"
	"github.com/saveio/dsp-go-sdk/consts"
	tkComm "github.com/saveio/scan/tracker/common"
	theComm "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

// Marshalled as binary by the UDP client, so be careful making changes.
type AnnounceRequest struct {
	InfoHash   [consts.PROTO_NODE_FILE_HASH_LEN]byte
	PeerId     [20]byte
	Downloaded int64
	Left       uint64
	Uploaded   int64
	// Apparently this is optional. None can be used for announces done at
	// regular intervals.
	Event     AnnounceEvent
	IPAddress [4]byte
	Key       int32
	NumWant   int32 // How many peer addresses are desired. -1 for default.
	Port      uint16
	Wallet    theComm.Address
	NodeType  NodeType
} // 82 bytes

type AnnounceRequestOptions struct {
	PubKeyTLV    tkComm.TLV
	SignatureTLV tkComm.TLV
}

type AnnounceResponse struct {
	Interval  int32 // Minimum seconds the local peer should wait before next announce.
	Leechers  int32
	Seeders   int32
	Peers     []Peer
	IPAddress [4]byte
	Port      uint16
	Wallet    [20]byte
	NodesInfo *NodesInfoSt
}

func (req *AnnounceRequest) Print() {
	nodeAddr := krpc.NodeAddr{IP: req.IPAddress[:], Port: int(req.Port)}
	addr, err := theComm.AddressParseFromBytes(req.Wallet[:])
	log.Debugf("AnnounceRequestPrint. InfoHash: %v, PeerId: %v, Event: %s\n", string(req.InfoHash[:]), req.PeerId, req.Event)
	if err != nil {
		log.Debugf("can not parse address from bytes, address: %v\n", req.Wallet)
		log.Debugf("AnnounceRequestPrint. Wallet: %v, HostPort: %s, Key: %d, NumWant: %d, NodeType: %d\n", req.Wallet, nodeAddr.String(), req.Key, req.NumWant, req.NodeType)
	} else {
		log.Debugf("AnnounceRequestPrint. Wallet: %s, HostPort: %s, Key: %d, NumWant: %d, NodeType: %d\n", addr.ToBase58(), nodeAddr.String(), req.Key, req.NumWant, req.NodeType)
	}
	log.Debugf("AnnounceRequestPrint. Downloaded: %d, Left: %d, Upoaded: %d\n", req.Downloaded, req.Left, req.Uploaded)
}

func (res *AnnounceResponse) Print() {
	nodeAddr := krpc.NodeAddr{IP: res.IPAddress[:], Port: int(res.Port)}
	addr, err := theComm.AddressParseFromBytes(res.Wallet[:])
	if err != nil {
		log.Debugf("can not parse address from bytes, address: %v\n", res.Wallet)
		log.Debugf("AnnounceResponsePrint. Interval: %d, Leechers: %d, Seeders: %d, HostPort: %s, Wallet: %v\n", res.Interval, res.Leechers, res.Seeders, nodeAddr.String(), res.Wallet)
	} else {
		log.Debugf("AnnounceResponsePrint. Interval: %d, Leechers: %d, Seeders: %d, HostPort: %s, Wallet: %s\n", res.Interval, res.Leechers, res.Seeders, nodeAddr.String(), addr.ToBase58())
	}

	for i, peer := range res.Peers {
		log.Debugf("AnnounceResponsePrint. Peer_%d: %s, ID: %v\n", i, krpc.NodeAddr{IP: peer.IP, Port: peer.Port}.String(), peer.ID)
	}
}

type AnnounceEvent int32

const (
	AnnounceEventEmpty AnnounceEvent = iota
	AnnounceEventCompleted
	AnnounceEventStarted
	AnnounceEventStopped
	AnnounceEventUpdated
)

func (e AnnounceEvent) String() string {
	// See BEP 3, "event".
	return []string{"empty", "completed", "started", "stopped", "updated"}[e]
}

const (
	None      AnnounceEvent = iota
	Completed               // The local peer just completed the torrent.
	Started                 // The local peer has just resumed this torrent.
	Stopped                 // The local peer is leaving the swarm.
)

var (
	ErrBadScheme = errors.New("unknown scheme")
)

type Announce struct {
	TrackerUrl     string
	Request        AnnounceRequest
	RequestOptions AnnounceRequestOptions
	HostHeader     string
	HTTPProxy      func(*http.Request) (*url.URL, error)
	ServerName     string
	UserAgent      string
	UdpNetwork     string
	// If the Port is zero, it's assumed to be the same as the Request.Port
	ClientIp4 krpc.NodeAddr
	// If the Port is zero, it's assumed to be the same as the Request.Port
	ClientIp6 krpc.NodeAddr
	flag      ActionFlag
}

// In an FP language with currying, what order what you put these params?
func (me Announce) Do() (res AnnounceResponse, err error) {
	_url, err := url.Parse(me.TrackerUrl)
	if err != nil {
		return
	}
	switch _url.Scheme {
	case "http", "https":
		return announceHTTP(me, _url)
	case "udp", "udp4", "udp6":
		return announceUDP(me, _url)
	default:
		err = ErrBadScheme
		return
	}
}
