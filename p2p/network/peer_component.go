package network

import (
	"github.com/saveio/carrier/network"
	"github.com/saveio/themis/common/log"
)

type PeerComponent struct {
	Net *Network
}

func (this *PeerComponent) Startup(net *network.Network) {
}

func (this *PeerComponent) Cleanup(net *network.Network) {
}

func (this *PeerComponent) Receive(ctx *network.ComponentContext) error {
	return nil
}

func (this *PeerComponent) PeerConnect(client *network.PeerClient) {
	if client == nil || len(client.Address) == 0 {
		log.Warnf("peer has connected, but client is nil", client)
		return
	}
	peerId := client.ClientID()
	walletAddr := this.Net.walletAddrFromPeerId(peerId)
	if !isValidWalletAddr(walletAddr) {
		log.Warnf("receive invalid wallet addr %s", walletAddr)
		return
	}
	hostAddr := client.Address
	p, ok := this.Net.peers.LoadOrStore(walletAddr, New(hostAddr))
	var pr *Peer
	if ok {
		pr, ok = p.(*Peer)
		if !ok {
			log.Errorf("convert peer to peer.Peer failed")
			return
		}
	}
	pr.SetPeerId(peerId)
	log.Debugf("Store wallet %s for peer %s %s", walletAddr, client.Address, peerId)
	this.Net.peers.Store(walletAddr, pr)
}

func (this *PeerComponent) PeerDisconnect(client *network.PeerClient) {
	if client == nil || len(client.Address) == 0 {
		log.Warnf("peer has disconnected, but its address is clean")
		return
	}
	peerId := client.ClientID()
	walletAddr := this.Net.walletAddrFromPeerId(peerId)
	log.Debugf("peer has disconnect %s", walletAddr)
	this.Net.peers.Delete(walletAddr)
}
