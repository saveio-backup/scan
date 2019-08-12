package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/anacrolix/dht/krpc"
	"github.com/syndtr/goleveldb/leveldb"
)

var EDB *EndpointDB

const (
	EP_WALLET_ADDR_KEY_PREFIX = "wallet_addr:"
)

type EndpointDB struct {
	db   *LevelDBStore
	lock sync.RWMutex
}

type Endpoint struct {
	WalletAddr string        `json:"wallet_addr"`
	NodeAddr   krpc.NodeAddr `json:"node_addr"`
}

func NewEndpointDB(db *LevelDBStore) *EndpointDB {
	return &EndpointDB{
		db: db,
	}
}

func (this *EndpointDB) Close() error {
	return this.db.Close()
}

func (this *EndpointDB) PutEndpoint(walletAddr string, host net.IP, port int) error {
	this.lock.Lock()
	defer this.lock.Unlock()

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
	fmt.Println(ep.NodeAddr.IP)
	buf, err := json.Marshal(ep)
	if err != nil {
		return err
	}
	return this.db.Put(key, buf)
}

func (this *EndpointDB) GetEndpoint(walletAddr string) (*Endpoint, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	value, err := this.db.Get([]byte(EP_WALLET_ADDR_KEY_PREFIX + walletAddr))
	if err != nil && err != leveldb.ErrNotFound {
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

func (this *EndpointDB) DelEndpoint(walletAddr string) error {
	return this.db.Delete([]byte(EP_WALLET_ADDR_KEY_PREFIX + walletAddr))
}
