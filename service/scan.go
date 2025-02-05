package service

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/saveio/dsp-go-sdk/consts"
	"github.com/saveio/scan/service/tk"

	dspCfg "github.com/saveio/dsp-go-sdk/config"
	"github.com/saveio/dsp-go-sdk/core/channel"
	"github.com/saveio/dsp-go-sdk/utils/crypto"
	putils "github.com/saveio/pylons/utils"
	"github.com/saveio/scan/common/config"
	ch_actor_server "github.com/saveio/scan/p2p/actor/channel/server"
	dns_actor_server "github.com/saveio/scan/p2p/actor/dns/server"
	tk_actor_client "github.com/saveio/scan/p2p/actor/tracker/client"
	tk_actor_server "github.com/saveio/scan/p2p/actor/tracker/server"
	"github.com/saveio/scan/p2p/network"
	themisSdk "github.com/saveio/themis-go-sdk"
	chainsdk "github.com/saveio/themis-go-sdk/utils"
	"github.com/saveio/themis/account"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/errors"
	dnsContract "github.com/saveio/themis/smartcontract/service/native/dns"
)

var ScanNode *Node

type Node struct {
	Config     *dspCfg.DspConfig
	Account    *account.Account
	Chain      *themisSdk.Chain
	DnsNet     *network.Network
	ChannelNet *network.Network
	TkNet      *network.Network
	Channel    *channel.Channel
	// Db         *storage.LevelDBStore
	PublicIp        string
	AccountPassword string
}

func Init(acc *account.Account, accPwd string) (*Node, error) {
	this := &Node{}
	this.Account = acc
	this.AccountPassword = accPwd
	config.SetCurrentUserWalletAddress(this.Account.Address.ToBase58())
	ScanNode = this
	return this, nil
}

func (this *Node) CurrentAccount() *account.Account {
	return this.Account
}

func (this *Node) StartScanNode(startChannelNetwork, startDnsNetwork, startTkNetwork bool) error {
	channelListenAddr := fmt.Sprintf("127.0.0.1:%d", int(config.Parameters.Base.PortBase+config.Parameters.Base.ChannelPortOffset))
	this.Config = &dspCfg.DspConfig{
		DBPath:               config.DspDBPath(),
		ChainRpcAddr:         config.Parameters.Base.ChainRpcAddr,
		ChannelClientType:    config.Parameters.Base.ChannelClientType,
		ChannelListenAddr:    channelListenAddr,
		ChannelProtocol:      config.Parameters.Base.ChannelProtocol,
		ChannelRevealTimeout: config.Parameters.Base.ChannelRevealTimeout,
		ChannelSettleTimeout: config.Parameters.Base.ChannelSettleTimeout,
		ChannelDBPath:        config.ChannelDBPath(),
		Mode:                 config.Parameters.Base.Mode,
	}

	this.Chain = themisSdk.NewChain()
	switch this.Config.Mode {
	case consts.DspModeOp:
		this.Chain.NewEthClient().SetAddress([]string{config.Parameters.Base.ChainRpcAddr})
		this.Chain.GetEthClient().DialClient()
	default:
		this.Chain.NewRpcClient().SetAddress([]string{config.Parameters.Base.ChainRpcAddr})
	}
	this.Chain.SetDefaultAccount(this.Account)

	if startChannelNetwork {
		err := this.SetupChannelNetwork()
		if err != nil {
			return err
		}
		log.Debugf("scan public addr is: %s\n", this.ChannelNet.PublicAddr())
		this.Channel, err = NewScanChannel(this, this.ChannelNet.GetPID())
		if err != nil {
			return err
		}

		err = this.RegOtherDnsEndpointsToSelf()
		if err != nil {
			return err
		}

		err = this.RegSelfEndpointToOtherDns()
		if err != nil {
			return err
		}

		if config.Parameters.Base.AutoSetupDNSRegisterEnable {
			err = this.autoRegisterDns()
			if err != nil {
				return err
			}
		}
		go this.connectAllDNSChannel()
	} else {
		// Auto register dns node as tracker even if channel network offline
		if err := this.autoRegisterDns(); err != nil {
			return err
		}
	}

	if startDnsNetwork {
		err := this.SetupDnsNetwork()
		if err != nil {
			return err
		}
	}

	if startTkNetwork {
		err := this.SetupTkNetwork()
		if err != nil {
			return err
		}
	}

	log.Info("scan node start success.")
	return nil
}

func (this *Node) SetupChannelNetwork() error {
	chActServer, err := ch_actor_server.NewP2PActor()
	if err != nil {
		return err
	}
	opts := []network.NetworkOption{
		network.WithKeys(crypto.NewNetworkKeyPairWithAccount(this.Account)),
		network.WithNetworkId(config.Parameters.Base.ChannelNetworkId),
		network.WithWalletAddrFromPeerId(crypto.AddressFromPubkeyHex),
		network.WithIntranetIP(config.Parameters.Base.IntranetIP),
		network.WithOpcodes(network.ChannelOpCodes),
		network.WithPid(chActServer.GetLocalPID()),
	}
	this.ChannelNet = network.NewP2P(opts...)
	chActServer.SetNetwork(this.ChannelNet)
	chActServer.SetDNSHostAddrFromWallet(this.GetDNSHostAddrFromWallet)

	return this.ChannelNet.Start(config.Parameters.Base.ChannelProtocol, config.Parameters.Base.PublicIP,
		fmt.Sprintf("%d", int(config.Parameters.Base.PortBase+config.Parameters.Base.ChannelPortOffset)))
}

func (this *Node) SetupDnsNetwork() error {
	log.Debugf("setup dns network")
	dnsActServer, err := dns_actor_server.NewP2PActor()
	if err != nil {
		return err
	}

	opts := []network.NetworkOption{
		network.WithKeys(crypto.NewNetworkKeyPairWithAccount(this.Account)),
		network.WithNetworkId(config.Parameters.Base.DnsNetworkId),
		network.WithWalletAddrFromPeerId(crypto.AddressFromPubkeyHex),
		network.WithIntranetIP(config.Parameters.Base.IntranetIP),
		network.WithOpcodes(network.TorrentOpCodes),
		network.WithPid(dnsActServer.GetLocalPID()),
	}
	this.DnsNet = network.NewP2P(opts...)
	network.DnsP2p = this.DnsNet
	dnsActServer.SetNetwork(this.DnsNet)

	port := fmt.Sprintf("%d", int(config.Parameters.Base.PortBase+config.Parameters.Base.DnsPortOffset))
	err = this.DnsNet.Start(config.Parameters.Base.DnsProtocol, config.Parameters.Base.PublicIP, port)
	if err != nil {
		return err
	}
	log.Debugf("setup dns network success with port %s", port)
	go func() {
		if err := this.SendConnectMsgToAllDns(); err != nil {
			log.Errorf("send connect msg to all dns err %s", err)
		}
	}()
	return nil
}

func (this *Node) SetupTkNetwork() error {
	log.Debugf("setup tk network")
	tkSrv := tk.NewTrackerService(this.DnsNet.GetPID(), this.Account.PublicKey, func(raw []byte) ([]byte, error) {
		return chainsdk.Sign(this.Account, raw)
	})
	tkSrv.SetCheckWhiteListFn(this.InWhiteListOrNot)
	tkActServer, err := tk_actor_server.NewTrackerActor(tkSrv)
	if err != nil {
		return err
	}

	opts := []network.NetworkOption{
		network.WithKeys(crypto.NewNetworkKeyPairWithAccount(this.Account)),
		network.WithNetworkId(config.Parameters.Base.TrackerNetworkId),
		network.WithWalletAddrFromPeerId(crypto.AddressFromPubkeyHex),
		network.WithIntranetIP(config.Parameters.Base.IntranetIP),
		network.WithOpcodes(network.TrackerOpCodes),
		network.WithPid(tkActServer.GetLocalPID()),
	}
	this.TkNet = network.NewP2P(opts...)
	network.TkP2p = this.TkNet
	tkActServer.SetNetwork(this.TkNet)
	tk_actor_client.SetTrackerServerPid(tkActServer.GetLocalPID())
	err = this.TkNet.Start(config.Parameters.Base.TrackerProtocol, config.Parameters.Base.PublicIP,
		fmt.Sprintf("%d", int(config.Parameters.Base.PortBase+config.Parameters.Base.TrackerPortOffset)))
	if err != nil {
		return err
	}
	log.Infof("tk public ip is %s", this.TkNet.PublicAddr())
	return nil
}

func (this *Node) SendConnectMsgToAllDns() error {
	allDns, err := this.GetAllDnsNodes()
	if err != nil {
		return err
	}

	for _, dns := range allDns {
		if dns.WalletAddr.ToBase58() == this.Account.Address.ToBase58() {
			continue
		}
		port, _ := strconv.Atoi(string(dns.Port))
		host := fmt.Sprintf("%s://%s:%d", config.Parameters.Base.DnsProtocol, dns.IP, port)
		log.Debugf("connect dns: wallet %s, host: %s", dns.WalletAddr.ToBase58(), host)
		_, err := this.ChannelNet.Connect(host)
		if err != nil {
			log.Warnf("connect dns error: %s", err)
		}
	}
	return nil
}

func (this *Node) autoRegisterDns() error {

	publicAddr := config.Parameters.Base.PublicIP
	if len(publicAddr) == 0 {
		publicAddr = this.ChannelNet.PublicAddr()
	}
	index := strings.Index(publicAddr, "://")
	hostPort := publicAddr
	if index != -1 {
		hostPort = publicAddr[index+3:]
	}
	host := publicAddr
	port := "10338"
	var err error
	if strings.Contains(hostPort, ":") {

		host, port, err = net.SplitHostPort(hostPort)
		if err != nil {
			return err
		}
	}

	queryWalletAddr := this.Account.Address
	var balance uint64
	switch this.Config.Mode {
	case consts.DspModeOp:
		balance, err = this.Chain.EVM.ERC20.BalanceOf(this.Account.EthAddress)
		queryWalletAddr, _ = common.AddressParseFromBytes(this.Account.EthAddress.Bytes())
	default:
		balance, err = this.Chain.Native.Usdt.BalanceOf(this.Account.Address)
	}
	log.Infof("auto register dns mode %s, %v, deposit %d", this.Config.Mode, balance, config.Parameters.Base.DnsGovernDeposit)
	if err != nil || balance < config.Parameters.Base.DnsGovernDeposit {
		log.Errorf("get dns balance: %s, governdeposit: %s, needs err: %v", putils.FormatUSDT(balance), putils.FormatUSDT(config.Parameters.Base.DnsGovernDeposit), err)
		return nil
	}

	ownNode, err := this.GetDnsNodeByAddr(queryWalletAddr)
	log.Debugf("ownNode: %v, err: %v\n", ownNode, err)

	if ownNode != nil && ownNode.WalletAddr.ToBase58() == queryWalletAddr.ToBase58() {
		if string(ownNode.IP) == host && string(ownNode.Port) == port {
			log.Infof("scan node info addr %s:%s not changed.", host, port)
			return nil
		}
		if _, err = this.DNSNodeUpdate(host, port); err != nil {
			return err
		}
	} else {
		if _, err = this.DNSNodeReg(host, port, config.Parameters.Base.DnsGovernDeposit); err != nil {
			return err
		}
	}

	log.Infof("scan node update addr to %s:%s success.", host, port)
	return nil
}

func (this *Node) AutoSetupDNSChannelsWorking() error {
	progress, err := this.GetFilterBlockProgress()
	if err != nil {
		return err
	}

	if progress.Progress != 1.0 {
		return errors.NewErr("block sync uncomplete, please wait a moment.")
	}

	item, err := this.GetDnsPeerPoolItem(hex.EncodeToString(keypair.SerializePublicKey(this.Account.PublicKey)))
	if err != nil {
		return err
	}

	if item.Status != dnsContract.RegisterCandidateStatus {
		return errors.NewErr("dns status is not RegisterCandidateStatus")
	}

	allDns, err := this.GetAllDnsNodes()
	if err != nil || len(allDns) == 0 {
		return err
	}

	allChannels, err := this.GetAllChannels()
	if err != nil {
		return err
	}

	setDNSNodeFunc := func(dnsUrl, walletAddr string) error {
		log.Debugf("dnsUrl: %s, walletAddr: %s\n", dnsUrl, walletAddr)
		if !isValidUrl(dnsUrl) {
			return errors.NewErr("invalid dnsUrl")
		}

		// open new channel
		channelId, err := this.Channel.OpenChannel(walletAddr, config.Parameters.Base.DnsChannelDeposit)
		if err != nil {
			return err
		}
		log.Debugf("auto setup dns channel success, channelId %d\n", channelId)

		bal, err := this.Channel.GetAvailableBalance(walletAddr)
		if err != nil {
			return err
		}
		log.Debugf("available balance %d\n", bal)
		return nil
	}

	numDnsChannels := 0
	for _, dns := range allDns {
		var dnsWalletAddrStr = dns.WalletAddr.ToBase58()
		var dnsUrlStr = fmt.Sprintf("%s://%s:%s", config.Parameters.Base.ChannelProtocol, dns.IP, dns.Port)

		// ignore to connect self, config define, channel is already exist.
		if dnsWalletAddrStr == this.Account.Address.ToBase58() ||
			contains(config.Parameters.Base.IgnoreConnectDNSAddrs, dnsWalletAddrStr) ||
			this.channelExists(allChannels.Channels, dnsWalletAddrStr) {
			continue
		}
		if numDnsChannels >= MAX_DNS_CHANNELS_NUM_AUTO_OPEN_WITH {
			break
		}
		err = setDNSNodeFunc(dnsUrlStr, dnsWalletAddrStr)
		if err != nil {
			log.Error(err)
		}
		numDnsChannels++
	}
	return nil
}

func (this *Node) connectAllDNSChannel() {
	max := 0
	for {
		<-time.After(time.Duration(10) * time.Second)
		max++
		if max > 12 {
			break
		}
		progress, err := this.GetFilterBlockProgress()
		if progress.Progress != 1.0 || err != nil {
			log.Debugf("channel is syncing")
			continue
		}
		allChannels, err := this.GetAllChannels()
		if err != nil {
			continue
		}
		if allChannels == nil {
			log.Debugf("no channel to connect when startup")
			continue
		}
		for _, ch := range allChannels.Channels {
			hostAddr := ch.HostAddr
			if len(ch.HostAddr) == 0 {
				log.Warnf("channel %s host addr is empty", ch.Address)
				publicIP, err := GetExternalIP(ch.Address)
				if err != nil || len(publicIP) == 0 {
					log.Errorf("get external ip for %s failed %s", ch.Address, err)
					continue
				}
				hostAddr = publicIP
			}
			log.Debugf("connect to dns start, hostAddr:%s, walletAddr %s", hostAddr, ch.Address)
			_, err := this.ChannelNet.Connect(hostAddr)
			if err != nil {
				log.Errorf("connect to dns failed, hostAddr: %s, err: %s", hostAddr, err)
				continue
			}
			log.Debugf("connect to dns success, hostAddr: %s, walletAddr: %s", hostAddr, ch.Address)
		}
		break
	}
}

var startChannelHeight uint32

type FilterBlockProgress struct {
	Progress float32
	Start    uint32
	End      uint32
	Now      uint32
}

func (this *Node) GetFilterBlockProgress() (*FilterBlockProgress, error) {
	progress := &FilterBlockProgress{}
	if this.Channel == nil {
		return progress, nil
	}
	endChannelHeight, err := this.Chain.GetCurrentBlockHeight()
	if err != nil {
		log.Debugf("get channel err %s", err)
		return progress, err
	}
	if endChannelHeight == 0 {
		return progress, nil
	}
	progress.Start = startChannelHeight
	progress.End = endChannelHeight
	now := this.Channel.GetCurrentFilterBlockHeight()
	progress.Now = now
	log.Debugf("endChannelHeight %d, start %d", endChannelHeight, startChannelHeight)
	if endChannelHeight <= startChannelHeight {
		progress.Progress = 1.0
		return progress, nil
	}
	rangeHeight := endChannelHeight - startChannelHeight
	if now >= rangeHeight+startChannelHeight {
		progress.Progress = 1.0
		return progress, nil
	}
	p := float32(now-startChannelHeight) / float32(rangeHeight)
	progress.Progress = p
	log.Debugf("GetFilterBlockProgress start %d, now %d, end %d, progress %v", startChannelHeight, now, endChannelHeight, progress)
	return progress, nil
}

func (this *Node) GetDNSHostAddrFromWallet(wallet string) string {
	hostAddr, _ := GetExternalIP(wallet)
	if len(hostAddr) > 0 {
		return hostAddr
	}

	walletAddress, err := common.AddressFromBase58(wallet)
	if err != nil {
		return ""
	}
	info, _ := this.GetDnsNodeByAddr(walletAddress)
	if info == nil {
		return ""
	}
	return fmt.Sprintf("%s://%s:%s", config.Parameters.Base.ChannelProtocol, info.IP, info.Port)
}
