package tracker

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/url"
	"time"

	"github.com/anacrolix/dht/krpc"
	"github.com/anacrolix/missinggo"
	"github.com/anacrolix/missinggo/pproffd"
	theComm "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

type Action int32
type ActionFlag = Action

const (
	ActionConnect Action = iota
	ActionAnnounce
	ActionScrape
	ActionError
	ActionReg
	ActionUnReg
	ActionReq
	ActionRegNodeType
	ActionGetNodesByType
	connectRequestConnectionId = 0x41727101980

	// BEP 41
	optionTypeEndOfOptions = 0
	optionTypeNOP          = 1
	optionTypeURLData      = 2

	// UDP timewait
	UDP_TIMEWAIT    = 15
	UDP_RETRY_TIMES = 3
)

type ConnectionRequest struct {
	ConnectionId int64
	Action       int32
	TransctionId int32
}

type ConnectionResponse struct {
	ConnectionId int64
}

type ResponseHeader struct {
	Action        Action
	TransactionId int32
}

type RequestHeader struct {
	ConnectionId  int64
	Action        Action
	TransactionId int32
} // 16 bytes

type AnnounceResponseHeader struct {
	Interval  int32
	Leechers  int32
	Seeders   int32
	IPAddress [4]byte
	Port      uint16
	Wallet    [20]byte
}

func (h *AnnounceResponseHeader) Print() {
	nodeAddr := krpc.NodeAddr{IP: h.IPAddress[:], Port: int(h.Port)}
	addr, err := theComm.AddressParseFromBytes(h.Wallet[:])
	if err != nil {
		log.Debugf("can not parse address from bytes, address: %v\n", h.Wallet)
		log.Debugf("AnnounceResponseHeaderPrint. Interval: %d, Leechers: %d, Seeders: %d, HostPort: %s, Wallet: %v\n", h.Interval, h.Leechers, h.Seeders, nodeAddr.String(), h.Wallet)
	} else {
		log.Debugf("AnnounceResponseHeaderPrint. Interval: %d, Leechers: %d, Seeders: %d, HostPort: %s, Wallet: %s\n", h.Interval, h.Leechers, h.Seeders, nodeAddr.String(), addr.ToBase58())
	}
}

func newTransactionId() int32 {
	return int32(rand.Uint32())
}

func timeout(contiguousTimeouts int) (d time.Duration) {
	if contiguousTimeouts > 3 {
		contiguousTimeouts = 3
	}
	d = UDP_TIMEWAIT * time.Second
	for ; contiguousTimeouts > 0; contiguousTimeouts-- {
		d *= 2
	}
	return
}

type udpAnnounce struct {
	contiguousTimeouts   int
	connectionIdReceived time.Time
	connectionId         int64
	socket               net.Conn
	url                  url.URL
	a                    *Announce
}

func (c *udpAnnounce) Close() error {
	if c.socket != nil {
		return c.socket.Close()
	}
	return nil
}

func (c *udpAnnounce) ipv6() bool {
	if c.a.UdpNetwork == "udp6" {
		return true
	}
	rip := missinggo.AddrIP(c.socket.RemoteAddr())
	return rip.To16() != nil && rip.To4() == nil
}

func SerializeReqOptionsToBytes(reqOptions AnnounceRequestOptions) (options []byte, err error) {
	var pubKeyBuf, signBuf bytes.Buffer
	if reqOptions.PubKeyTLV != nil {
		err = reqOptions.PubKeyTLV.Write(&pubKeyBuf)
		if err != nil {
			log.Error(err)
		}
		options = append(options, pubKeyBuf.Bytes()...)
	}
	if reqOptions.SignatureTLV != nil {
		err = reqOptions.SignatureTLV.Write(&signBuf)
		if err != nil {
			log.Error(err)
		}
		options = append(options, signBuf.Bytes()...)
	}
	return options, err
}

func (c *udpAnnounce) Do(req AnnounceRequest, reqOptions AnnounceRequestOptions) (res AnnounceResponse, err error) {
	err = c.connect()
	if err != nil {
		log.Error("Do connect err: ", err.Error())
		return AnnounceResponse{}, err
	}

	req.Print()
	// reqURI := c.url.RequestURI()
	//if c.ipv6() {
	//	// BEP 15
	//	req.IPAddress = [4]byte{0x0}
	//} else if req.IPAddress == [4]byte{0x0} && c.a.ClientIp4.IP != nil {
	//	log.Warnf("in c.ipv6 elif,req.ip:%v",req.IPAddress)
	//	req.IPAddress = ipconvert(c.a.ClientIp4.IP.To4())
	//}
	// Clearly this limits the request URI to 255 bytes. BEP 41 supports
	// longer but I'm not fussed.
	// options := append([]byte{optionTypeURLData, byte(len(reqURI))}, []byte(reqURI)...)
	options, err := SerializeReqOptionsToBytes(reqOptions)
	var b *bytes.Buffer
	flag := c.a.flag
	if flag != 0 {
		b, err = c.request(flag, req, options)
	} else {
		b, err = c.request(ActionAnnounce, req, options)
	}
	if err != nil {
		return AnnounceResponse{}, err
	}
	var h AnnounceResponseHeader
	err = readBody(b, &h)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		err = fmt.Errorf("error parsing announce response: %s", err)
		return AnnounceResponse{}, err
	}
	h.Print()

	if c.a.flag == ActionAnnounce {
		nas := func() interface {
			encoding.BinaryUnmarshaler
			NodeAddrs() []krpc.NodeAddr
		} {
			if c.ipv6() {
				return &krpc.CompactIPv6NodeAddrs{}
			} else {
				return &krpc.CompactIPv4NodeAddrs{}
			}
		}()
		err = nas.UnmarshalBinary(b.Bytes())
		if err != nil {
			return AnnounceResponse{}, err
		}
		for _, cp := range nas.NodeAddrs() {
			res.Peers = append(res.Peers, Peer{}.FromNodeAddr(cp))
		}
	} else if c.a.flag == ActionGetNodesByType {
		log.Debugf("Client ActionGetNodesByType: %v", b.Bytes())
		res.NodesInfo = &NodesInfoSt{}
		err = res.NodesInfo.DeSerialize(b)
	}

	res.Interval, res.Leechers, res.Seeders, res.IPAddress, res.Port, res.Wallet = h.Interval, h.Leechers, h.Seeders, h.IPAddress, h.Port, h.Wallet
	res.Print()
	return res, nil
}

// body is the binary serializable request body. trailer is optional data
// following it, such as for BEP 41.
func (c *udpAnnounce) write(h *RequestHeader, body interface{}, trailer []byte) (err error) {
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, h)
	if err != nil {
		panic(err)
	}
	if body != nil {
		err = binary.Write(&buf, binary.BigEndian, body)
		if err != nil {
			panic(err)
		}
	}
	_, err = buf.Write(trailer)
	if err != nil {
		return
	}
	n, err := c.socket.Write(buf.Bytes())
	if err != nil {
		return
	}
	if n != buf.Len() {
		panic("write should send all or error")
	}
	return
}

func read(r io.Reader, data interface{}) error {
	return binary.Read(r, binary.BigEndian, data)
}

func write(w io.Writer, data interface{}) error {
	return binary.Write(w, binary.BigEndian, data)
}

// args is the binary serializable request body. trailer is optional data
// following it, such as for BEP 41.
func (c *udpAnnounce) request(action Action, args interface{}, options []byte) (responseBody *bytes.Buffer, err error) {
	tid := newTransactionId()
	err = c.write(&RequestHeader{
		ConnectionId:  c.connectionId,
		Action:        action,
		TransactionId: tid,
	}, args, options)
	if err != nil {
		log.Errorf("tracker.udp.request write err: %v\n", err)
		return
	}

	c.socket.SetReadDeadline(time.Now().Add(timeout(c.contiguousTimeouts)))
	b := make([]byte, 0x800) // 2KiB
	for {
		var n int
		n, err = c.socket.Read(b)
		if err != nil {
			log.Errorf("tracker.udp.request.socket.Read err: %v\n", err)
			return
		}

		buf := bytes.NewBuffer(b[:n])
		var h ResponseHeader
		err = binary.Read(buf, binary.BigEndian, &h)
		switch err {
		case io.ErrUnexpectedEOF:
			log.Errorf("tracker.udp.request.binary.Read err.ErrUnexpectedEOF: %v\n", err)
			continue
		case nil:
		default:
			log.Errorf("tracker.udp.request.binary.Read err.default: %v\n", err)
			return
		}
		if h.TransactionId != tid {
			log.Debugf("tracker.udp.request.binary.Read h.TransactionId != tid: %d, %d\n", h.TransactionId, tid)
			continue
		}
		c.contiguousTimeouts = 0
		if h.Action == ActionError {
			err = errors.New(buf.String())
			log.Debugf("tracker.udp.request.return h.Action == ActionError err: %v\n", err)
		}
		responseBody = buf
		return
	}
}

func readBody(r io.Reader, data ...interface{}) (err error) {
	for _, datum := range data {
		err = binary.Read(r, binary.BigEndian, datum)
		if err != nil {
			break
		}
	}
	return
}

func (c *udpAnnounce) connected() bool {
	return !c.connectionIdReceived.IsZero() && time.Now().Before(c.connectionIdReceived.Add(time.Minute))
}

func (c *udpAnnounce) dialNetwork() string {
	if c.a.UdpNetwork != "" {
		return c.a.UdpNetwork
	}
	return "udp"
}

func (c *udpAnnounce) connect() (err error) {
	if c.connected() {
		log.Debugf("tracker.udp.connect.connected")
		return nil
	}
	c.connectionId = connectRequestConnectionId
	if c.socket == nil {
		hmp := missinggo.SplitHostMaybePort(c.url.Host)
		if hmp.NoPort {
			hmp.NoPort = false
			hmp.Port = 80
		}
		c.socket, err = net.Dial(c.dialNetwork(), hmp.String())
		if err != nil {
			log.Debugf("Net Dial error:%v", err)
			return err
		}
		log.Debugf("Net Dial success,socket:%v", c.socket)
		c.socket = pproffd.WrapNetConn(c.socket)
	}
	b, err := c.request(ActionConnect, nil, nil)
	if err != nil {
		log.Error(fmt.Sprintf("tracker.udp.connect.c.request err: %v", err))
		return err
	}
	var res ConnectionResponse
	err = readBody(b, &res)
	if err != nil {
		log.Error(fmt.Sprintf("tracker.udp.connect.readBody err: %v", err))
		return err
	}
	c.connectionId = res.ConnectionId
	c.connectionIdReceived = time.Now()
	return
}

// TODO: Split on IPv6, as BEP 15 says response peer decoding depends on
// network in use.
func announceUDP(opt Announce, _url *url.URL) (AnnounceResponse, error) {
	ua := udpAnnounce{
		url: *_url,
		a:   &opt,
	}
	defer ua.Close()

	retryTimes := 0
	var ret AnnounceResponse
	var err error

	for retryTimes < UDP_RETRY_TIMES {
		ret, err = ua.Do(opt.Request, opt.RequestOptions)
		if err != nil {
			log.Errorf(" announceUDP err: %v\n", err)
			if opE, ok := err.(*net.OpError); ok {
				log.Errorf(" announceUDP err opE: %v\n", opE)
				if opE.Timeout() {
					retryTimes++
					log.Errorf("announceUDP timeout retryTimes: %d \n", retryTimes)
					continue
				}
			}
			return ret, err
		}
		return ret, err
	}

	return ret, errors.New(fmt.Sprintf("udp retry %d times, also failed. err: %v", retryTimes, err))
}
