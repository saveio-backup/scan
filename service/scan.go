package service

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/saveio/carrier/crypto"
	"github.com/saveio/carrier/crypto/ed25519"

	dspCfg "github.com/saveio/dsp-go-sdk/config"
	"github.com/saveio/dsp-go-sdk/core/channel"
	"github.com/saveio/scan/common/config"
	ch_actor_server "github.com/saveio/scan/p2p/actor/channel/server"
	dns_actor_server "github.com/saveio/scan/p2p/actor/dns/server"
	tk_actor_client "github.com/saveio/scan/p2p/actor/tracker/client"
	tk_actor_server "github.com/saveio/scan/p2p/actor/tracker/server"
	channel_net "github.com/saveio/scan/p2p/networks/channel"
	dns_net "github.com/saveio/scan/p2p/networks/dns"
	tk_net "github.com/saveio/scan/p2p/networks/tracker"
	tk "github.com/saveio/scan/service/tk"
	themisSdk "github.com/saveio/themis-go-sdk"
	chainsdk "github.com/saveio/themis-go-sdk/utils"
	"github.com/saveio/themis/account"
	cutils "github.com/saveio/themis/cmd/utils"
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
	DnsNet     *dns_net.Network
	ChannelNet *channel_net.Network
	TkNet      *tk_net.Network
	Channel    *channel.Channel
	// Db         *storage.LevelDBStore
	PublicIp string
}

func Init(acc *account.Account) (*Node, error) {
	this := &Node{}
	this.Account = acc
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
	}

	this.Chain = themisSdk.NewChain()
	this.Chain.NewRpcClient().SetAddress([]string{config.Parameters.Base.ChainRpcAddr})
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

	cPub := keypair.SerializePublicKey(this.Account.PubKey())
	chPub, chPri, err := ed25519.GenerateKey(&accountReader{
		PublicKey: append(cPub, []byte("channel")...),
	})
	chNetworkKey := &crypto.KeyPair{
		PublicKey:  chPub,
		PrivateKey: chPri,
	}
	this.ChannelNet = channel_net.NewP2P()
	this.ChannelNet.SetNetworkKey(chNetworkKey)
	this.ChannelNet.SetProxyServer(config.Parameters.Base.NATProxyServerAddr)
	this.ChannelNet.SetPID(chActServer.GetLocalPID())
	chActServer.SetNetwork(this.ChannelNet)

	return this.ChannelNet.Start(config.Parameters.Base.ChannelProtocol, config.Parameters.Base.PublicIP,
		fmt.Sprintf("%d", int(config.Parameters.Base.PortBase+config.Parameters.Base.ChannelPortOffset)))
}

func (this *Node) SetupDnsNetwork() error {
	log.Debugf("setup dns network")
	dnsActServer, err := dns_actor_server.NewP2PActor()
	if err != nil {
		return err
	}

	dPub := keypair.SerializePublicKey(this.Account.PubKey())
	dnsPub, dnsPri, err := ed25519.GenerateKey(&accountReader{
		PublicKey: append(dPub, []byte("dns")...),
	})
	dnsNetworkKey := &crypto.KeyPair{
		PublicKey:  dnsPub,
		PrivateKey: dnsPri,
	}
	this.DnsNet = dns_net.NewP2P()
	this.DnsNet.SetNetworkKey(dnsNetworkKey)
	this.DnsNet.SetProxyServer(config.Parameters.Base.NATProxyServerAddr)
	this.DnsNet.SetPID(dnsActServer.GetLocalPID())

	dns_net.DnsP2p = this.DnsNet
	dnsActServer.SetNetwork(this.DnsNet)

	err = this.DnsNet.Start(config.Parameters.Base.DnsProtocol, config.Parameters.Base.PublicIP,
		fmt.Sprintf("%d", int(config.Parameters.Base.PortBase+config.Parameters.Base.DnsPortOffset)))
	if err != nil {
		return err
	}
	log.Debugf("setup dns network success")
	return this.SendConnectMsgToAllDns()
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

	dPub := keypair.SerializePublicKey(this.Account.PubKey())
	tkPub, tkPri, err := ed25519.GenerateKey(&accountReader{
		PublicKey: append(dPub, []byte("tk")...),
	})
	tkNetworkKey := &crypto.KeyPair{
		PublicKey:  tkPub,
		PrivateKey: tkPri,
	}
	this.TkNet = tk_net.NewP2P()
	this.TkNet.SetNetworkKey(tkNetworkKey)
	this.TkNet.SetProxyServer(config.Parameters.Base.NATProxyServerAddr)
	this.TkNet.SetPID(tkActServer.GetLocalPID())
	tk_net.TkP2p = this.TkNet
	tkActServer.SetNetwork(this.TkNet)
	tk_actor_client.SetTrackerServerPid(tkActServer.GetLocalPID())

	err = this.TkNet.Start(config.Parameters.Base.TrackerProtocol, config.Parameters.Base.PublicIP,
		fmt.Sprintf("%d", int(config.Parameters.Base.PortBase+config.Parameters.Base.TrackerPortOffset)),
		config.Parameters.Base.TrackerNetworkId)
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
		dnsPortOffset := config.Parameters.Base.PortBase + config.Parameters.Base.DnsPortOffset
		log.Debugf("connect dns: %s\n", fmt.Sprintf("%s://%s:%d", config.Parameters.Base.DnsProtocol, dns.IP, dnsPortOffset))
		err := this.DnsNet.ConnectAndWait(fmt.Sprintf("%s://%s:%d", config.Parameters.Base.DnsProtocol, dns.IP, dnsPortOffset))
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func (this *Node) autoRegisterDns() error {
	publicAddr := this.ChannelNet.PublicAddr()
	index := strings.Index(publicAddr, "://")
	hostPort := publicAddr
	if index != -1 {
		hostPort = publicAddr[index+3:]
	}
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return err
	}
	balance, err := this.Chain.Native.Usdt.BalanceOf(this.Account.Address)
	if err != nil || balance < config.Parameters.Base.DnsGovernDeposit {
		log.Errorf("get dns balance: %s, governdeposit: %s, needs err: %v", cutils.FormatUsdt(balance), cutils.FormatUsdt(config.Parameters.Base.DnsGovernDeposit), err)
		return nil
	}
	ownNode, err := this.GetDnsNodeByAddr(this.Account.Address)
	log.Debugf("ownNode: %v, err: %v\n", ownNode, err)

	if ownNode != nil {
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
