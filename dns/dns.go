package dns

import (
	"fmt"

	"github.com/saveio/scan/common/config"
	themisSdk "github.com/saveio/themis-go-sdk"
	"github.com/saveio/themis/account"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/smartcontract/service/native/dns"
)

// 全局指针变量GlbDNSSvr可能存在数据竞争问题
var GlbDNSSvr *DNSSvr

type DNSSvr struct {
	Account *account.Account
	Chain   *themisSdk.Chain
}

func NewDNSSvr(Account *account.Account) (*DNSSvr, error) {
	if GlbDNSSvr != nil {
		return GlbDNSSvr, nil
	}

	ds := &DNSSvr{
		Chain: themisSdk.NewChain(),
	}
	chainRpcAddr := fmt.Sprintf("http://127.0.0.1:%d", config.DefaultConfig.CommonConfig.ChainRpcAddr)
	ds.Chain.NewRpcClient().SetAddress([]string{chainRpcAddr})

	ds.Account = Account
	ds.Chain.SetDefaultAccount(ds.Account)

	log.Info("GenGlobalEndPoint start successed.")
	return ds, nil
}

//dns api
func (this *DNSSvr) DNSNodeReg(ip string, port string, initDeposit uint64) (common.Uint256, error) {
	return this.Chain.Native.Dns.DNSNodeReg([]byte(ip), []byte(port), initDeposit)
}

func (this *DNSSvr) DNSNodeUnreg() (common.Uint256, error) {
	return this.Chain.Native.Dns.UnregisterDNSNode()
}

func (this *DNSSvr) DNSNodeQuit() (common.Uint256, error) {
	return this.Chain.Native.Dns.QuitNode()
}

func (this *DNSSvr) DNSNodeUpdate(ip, port string) (common.Uint256, error) {
	return this.Chain.Native.Dns.UpdateNode([]byte(ip), []byte(port))
}

func (this *DNSSvr) DNSAddPos(deltaDeposit uint64) (common.Uint256, error) {
	return this.Chain.Native.Dns.AddInitPos(deltaDeposit)
}

func (this *DNSSvr) DNSReducePos(deltaDeposit uint64) (common.Uint256, error) {
	return this.Chain.Native.Dns.ReduceInitPos(deltaDeposit)
}

func (this *DNSSvr) GetDnsPeerPoolMap() (*dns.PeerPoolMap, error) {
	return this.Chain.Native.Dns.GetPeerPoolMap()
}

func (this *DNSSvr) GetDnsPeerPoolItem(pubKey string) (*dns.PeerPoolItem, error) {
	return this.Chain.Native.Dns.GetPeerPoolItem(pubKey)
}

func (this *DNSSvr) GetAllDnsNodes() (map[string]dns.DNSNodeInfo, error) {
	return this.Chain.Native.Dns.GetAllDnsNodes()
}

func (this *DNSSvr) GetDnsNodeByAddr(addr common.Address) (*dns.DNSNodeInfo, error) {
	return this.Chain.Native.Dns.GetDnsNodeByAddr(addr)
}
