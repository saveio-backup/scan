package service

import (
	"github.com/saveio/dsp-go-sdk/consts"
	sdkDns "github.com/saveio/themis-go-sdk/dns"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/smartcontract/service/native/dns"
)

func (this *Node) GetDns() *sdkDns.Dns {
	switch this.Config.Mode {
	case consts.DspModeOp:
		return this.Chain.EVM.Dns
	default:
		return this.Chain.Native.Dns
	}
}

func (this *Node) DNSNodeReg(ip string, port string, initDeposit uint64) (common.Uint256, error) {
	return this.GetDns().DNSNodeReg([]byte(ip), []byte(port), initDeposit)
}

func (this *Node) DNSNodeUnreg() (common.Uint256, error) {
	return this.GetDns().UnregisterDNSNode()
}

func (this *Node) DNSNodeQuit() (common.Uint256, error) {
	return this.GetDns().QuitNode()
}

func (this *Node) DNSNodeUpdate(ip, port string) (common.Uint256, error) {
	return this.GetDns().UpdateNode([]byte(ip), []byte(port))
}

func (this *Node) DNSAddPos(deltaDeposit uint64) (common.Uint256, error) {
	return this.GetDns().AddInitPos(deltaDeposit)
}

func (this *Node) DNSReducePos(deltaDeposit uint64) (common.Uint256, error) {
	return this.GetDns().ReduceInitPos(deltaDeposit)
}

func (this *Node) GetDnsPeerPoolMap() (*dns.PeerPoolMap, error) {
	return this.GetDns().GetPeerPoolMap()
}

func (this *Node) GetDnsPeerPoolItem(pubKey string) (*dns.PeerPoolItem, error) {
	return this.GetDns().GetPeerPoolItem(pubKey)
}

func (this *Node) GetAllDnsNodes() (map[string]dns.DNSNodeInfo, error) {
	return this.GetDns().GetAllDnsNodes()
}

func (this *Node) GetDnsNodeByAddr(addr common.Address) (*dns.DNSNodeInfo, error) {
	return this.GetDns().GetDnsNodeByAddr(addr)
}
