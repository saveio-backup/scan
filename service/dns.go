package service

import (
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/dns"
)

//dns api
func (this *Node) DNSNodeReg(ip string, port string, initDeposit uint64) (common.Uint256, error) {
	return this.Chain.Native.Dns.DNSNodeReg([]byte(ip), []byte(port), initDeposit)
}

func (this *Node) DNSNodeUnreg() (common.Uint256, error) {
	return this.Chain.Native.Dns.UnregisterDNSNode()
}

func (this *Node) DNSNodeQuit() (common.Uint256, error) {
	return this.Chain.Native.Dns.QuitNode()
}

func (this *Node) DNSNodeUpdate(ip, port string) (common.Uint256, error) {
	return this.Chain.Native.Dns.UpdateNode([]byte(ip), []byte(port))
}

func (this *Node) DNSAddPos(deltaDeposit uint64) (common.Uint256, error) {
	return this.Chain.Native.Dns.AddInitPos(deltaDeposit)
}

func (this *Node) DNSReducePos(deltaDeposit uint64) (common.Uint256, error) {
	return this.Chain.Native.Dns.ReduceInitPos(deltaDeposit)
}

func (this *Node) GetDnsPeerPoolMap() (*dns.PeerPoolMap, error) {
	return this.Chain.Native.Dns.GetPeerPoolMap()
}

func (this *Node) GetDnsPeerPoolItem(pubKey string) (*dns.PeerPoolItem, error) {
	return this.Chain.Native.Dns.GetPeerPoolItem(pubKey)
}

func (this *Node) GetAllDnsNodes() (map[string]dns.DNSNodeInfo, error) {
	return this.Chain.Native.Dns.GetAllDnsNodes()
}

func (this *Node) GetDnsNodeByAddr(addr common.Address) (*dns.DNSNodeInfo, error) {
	return this.Chain.Native.Dns.GetDnsNodeByAddr(addr)
}
