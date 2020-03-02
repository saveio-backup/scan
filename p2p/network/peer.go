package network

import (
	"sync"

	"github.com/saveio/themis/common/log"
)

type Peer struct {
	id   string        // peer id
	addr string        // peer host address
	lock *sync.RWMutex // lock
}

func New(addr string) *Peer {
	p := &Peer{
		addr: addr,
		lock: new(sync.RWMutex),
	}
	return p
}

// GetHostAddr.
func (p *Peer) GetHostAddr() string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.addr
}

func (p *Peer) SetPeerId(peerId string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	log.Debugf("set peer id %s for wallet %s, p %p", peerId, p.addr, p)
	p.id = peerId
}

// GetPeerId.
func (p *Peer) GetPeerId() string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	log.Debugf("get peer id %s for wallet %s, p %p", p.id, p.addr, p)
	return p.id
}
