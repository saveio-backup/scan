package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/anacrolix/dht/krpc"
	"github.com/syndtr/goleveldb/leveldb"
)

var EDB *EndpointDB

const (
	EP_WALLET_ADDR_KEY_PREFIX = "wallet_addr:"
)

type EndpointDB struct {
	Db *LevelDBStore
}

type Endpoint struct {
	WalletAddr string        `json:"wallet_addr"`
	NodeAddr   krpc.NodeAddr `json:"node_addr"`
}

func NewEndpointDB(db *LevelDBStore) *EndpointDB {
	return &EndpointDB{
		Db: db,
	}
}

func (this *EndpointDB) Close() error {
	return this.Db.Close()
}

func (this *EndpointDB) PutEndpoint(walletAddr string, host net.IP, port int) error {
	if host == nil || port < 0 || port > 65535 {
		return errors.New(fmt.Sprintf("invalid address format %s:%d", string(host), port))
	}
	ep := &Endpoint{
		WalletAddr: walletAddr,
		NodeAddr: krpc.NodeAddr{
			IP:   host.To4(),
			Port: port,
		},
	}
	key := []byte(EP_WALLET_ADDR_KEY_PREFIX + ep.WalletAddr)
	buf, err := json.Marshal(ep)
	if err != nil {
		return err
	}
	return this.Db.Put(key, buf)
}

func (this *EndpointDB) GetEndpoint(walletAddr string) (*Endpoint, error) {
	value, err := this.Db.Get([]byte(EP_WALLET_ADDR_KEY_PREFIX + walletAddr))
	if err != nil && err != leveldb.ErrNotFound && err.Error() != "not found" {
		return nil, err
	}
	if len(value) == 0 {
		return nil, nil
	}

	info := &Endpoint{}
	err = json.Unmarshal(value, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (this *EndpointDB) UpdateEndpoint(walletAddr, hostAddr string) error {

	host, port, err := net.SplitHostPort(hostAddr)
	if err != nil {
		return err
	}
	netIp := net.ParseIP(host).To4()
	if netIp == nil {
		netIp = net.ParseIP(host).To16()
	}
	netPort, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	ep, err := this.GetEndpoint(walletAddr)
	if err != nil {
		return err
	} else if ep != nil && ep.NodeAddr.String() == hostAddr {
		return nil
	}
	return this.PutEndpoint(walletAddr, netIp, netPort)
}

func (this *EndpointDB) DelEndpoint(walletAddr string) error {
	return this.Db.Delete([]byte(EP_WALLET_ADDR_KEY_PREFIX + walletAddr))
}
