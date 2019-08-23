package service

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/saveio/carrier/crypto"
	"github.com/saveio/carrier/crypto/ed25519"

	"github.com/saveio/dsp-go-sdk/channel"
	dspCfg "github.com/saveio/dsp-go-sdk/config"
	"github.com/saveio/scan/common/config"
	"github.com/saveio/scan/netserver"
	ch_actor_server "github.com/saveio/scan/p2p/actor/channel/server"
	dns_actor_server "github.com/saveio/scan/p2p/actor/dns/server"
	channel_net "github.com/saveio/scan/p2p/networks/channel"
	dns_net "github.com/saveio/scan/p2p/networks/dns"
	themisSdk "github.com/saveio/themis-go-sdk"
	"github.com/saveio/themis/account"
	cutils "github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/errors"
)

var ScanNode *Node

type Node struct {
	Config     *dspCfg.DspConfig
	Account    *account.Account
	Chain      *themisSdk.Chain
	DnsNet     *dns_net.Network
	ChannelNet *channel_net.Network
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

func (this *Node) StartScanNode(startChannelNetwork, startDnsNetwork bool) error {
	channelListenAddr := fmt.Sprintf("127.0.0.1:%d", int(config.Parameters.Base.PortBase+config.Parameters.Base.ChannelPortOffset))
	this.Config = &dspCfg.DspConfig{
		DBPath:               config.DspDBPath(),
		ChainRpcAddr:         config.Parameters.Base.ChainRpcAddr,
		ChannelClientType:    config.Parameters.Base.ChannelClientType,
		ChannelListenAddr:    channelListenAddr,
		ChannelProtocol:      config.Parameters.Base.ChannelProtocol,
		ChannelRevealTimeout: config.Parameters.Base.ChannelRevealTimeout,
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
		this.Channel, err = NewScanChannel(this, this.ChannelNet.GetPID())
		if err != nil {
			return err
		}

		err = this.autoRegisterEndpoint()
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
	chListenAddr := fmt.Sprintf("%s://%s:%d",
		config.Parameters.Base.ChannelProtocol,
		config.Parameters.Base.PublicIP,
		int(config.Parameters.Base.PortBase+config.Parameters.Base.ChannelPortOffset))
	log.Debugf("goto start channel network %s", chListenAddr)
	chActServer.SetNetwork(this.ChannelNet)

	return this.ChannelNet.Start(chListenAddr)
}

func (this *Node) SetupDnsNetwork() error {
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
	dnsListenAddr := fmt.Sprintf("%s://%s:%d",
		config.Parameters.Base.DnsProtocol,
		config.Parameters.Base.PublicIP,
		int(config.Parameters.Base.PortBase+config.Parameters.Base.DnsPortOffset))
	log.Debugf("goto start dns network %s", dnsListenAddr)
	dns_net.DnsP2p = this.DnsNet
	dnsActServer.SetNetwork(this.DnsNet)

	return this.DnsNet.Start(dnsListenAddr)
}

func (this *Node) StartTrackerService() error {
	tkSvr := netserver.NewTKServer()
	tkSvr.Tsvr.SetPID(this.DnsNet.GetPID())
	go tkSvr.StartTrackerListening()
	log.Info("start tracker service success")
	return nil
}

func (this Node) autoRegisterEndpoint() error {
	publicAddr := this.ChannelNet.PublicAddr()
	index := strings.Index(publicAddr, "://")
	hostPort := publicAddr
	if index != -1 {
		hostPort = publicAddr[index+3:]
	}
	err := PutExternalIP(this.Account.Address.ToBase58(), hostPort)
	if err != nil {
		return err
	}

	ns, err := this.GetAllDnsNodes()
	if err != nil {
		return err
	}
	if len(ns) == 0 {
		return errors.NewErr("no dns nodes")
	}
	for _, v := range ns {
		err := PutExternalIP(v.WalletAddr.ToBase58(), fmt.Sprintf("%s:%s", string(v.IP), string(v.Port)))
		if err != nil {
			return err
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
		log.Fatal("get dns balance: %s, governdeposit: %s, needs err: %v", cutils.FormatUsdt(balance), cutils.FormatUsdt(config.Parameters.Base.DnsGovernDeposit), err)
	}
	ownNode, err := this.GetDnsNodeByAddr(this.Account.Address)
	if err != nil {
		log.Fatal("get dns node dnsinfo: %v, err: %v", ownNode, err)
	} else if ownNode != nil {
		_, err = this.DNSNodeUpdate(host, port)
	} else {
		_, err = this.DNSNodeReg(host, port, config.Parameters.Base.DnsGovernDeposit)
	}

	if err != nil {
		log.Fatalf("scan node update addr to %s:%s failed, err: %v", host, port, err)
		return err
	}
	log.Infof("scan node update addr to %s:%s success.", host, port)
	return nil
}

func (this *Node) AutoSetupDNSChannelsWorking() error {
	ns, err := this.GetAllDnsNodes()
	if err != nil {
		return err
	}
	if len(ns) == 0 {
		return errors.NewErr("no dns nodes")
	}

	setDNSNodeFunc := func(dnsUrl, walletAddr string) error {
		log.Debugf("set dns node func %s %s", dnsUrl, walletAddr)
		// if err := client.P2pIsPeerListening(dnsUrl); err != nil {
		// 	return err
		// }
		if strings.Index(dnsUrl, "0.0.0.0:0") != -1 {
			return errors.NewErr("invalid host addr")
		}
		// err = this.Channel.SetHostAddr(walletAddr, dnsUrl)
		// if err != nil {
		// 	return err
		// }
		_, err = this.Channel.OpenChannel(walletAddr, 0)
		if err != nil {
			log.Debugf("open channel err ")
			return err
		}
		err = this.Channel.WaitForConnected(walletAddr, time.Duration(100)*time.Second)
		if err != nil {
			log.Errorf("wait channel connected err %s %s", walletAddr, err)
			return err
		}
		log.Debugf("channel connected %s %s", walletAddr, err)
		bal, _ := this.Channel.GetAvailableBalance(walletAddr)
		log.Debugf("current balance %d", bal)
		log.Infof("connect to dns node :%s, deposit %d", dnsUrl, config.Parameters.Base.DnsChannelDeposit)
		err = this.Channel.SetDeposit(walletAddr, config.Parameters.Base.DnsChannelDeposit)
		if err != nil && strings.Index(err.Error(), "totalDeposit must big than contractBalance") == -1 {
			log.Debugf("deposit result %s", err)
			// TODO: withdraw and close channel
			return err
		}
		bal, _ = this.Channel.GetAvailableBalance(walletAddr)
		log.Debugf("current deposited balance %d", bal)
		log.Info("channel deposit success")

		err := this.Channel.CanTransfer(walletAddr, 10)
		if err == nil {
			log.Info("loopTest can transfer!")
			for i := 0; i < 3; i++ {
				err := this.Transfer(1, 1, walletAddr)
				if err != nil {
					log.Error("[loopTest] direct transfer failed:", err)
				} else {
					log.Info("[loopTest] direct transfer successfully")
				}
			}
		} else {
			if err != nil {
				log.Error("loopTest cannot transfer!")
			}
		}
		return nil
	}

	// first init
	for _, v := range ns {
		log.Debugf("auto setup dns channel working, dns walletaddr: %s hostaddr:%s:%s, ", v.WalletAddr.ToBase58(), string(v.IP), string(v.Port))

		// dnsUrl, _ := GetExternalIP(v.WalletAddr.ToBase58())
		var dnsUrl string
		dnsInfo, err := this.GetDnsNodeByAddr(v.WalletAddr)
		if err != nil || dnsInfo == nil {
			log.Errorf("auto setup dns channel working dnsinfo: %s, err: %s", dnsInfo, err)
		} else if len(dnsUrl) == 0 {
			dnsUrl = fmt.Sprintf("%s://%s:%s", config.Parameters.Base.ChannelProtocol, v.IP, v.Port)
		} else {
			dnsUrl = fmt.Sprintf("%s://%s:%s", config.Parameters.Base.ChannelProtocol, dnsInfo.IP, dnsInfo.Port)
		}
		err = PutExternalIP(v.WalletAddr.ToBase58(), strings.Split(dnsUrl, "://")[1])
		if err != nil {
			log.Errorf("auto setup dns channel working PutExternalIP err:%v", err)
		}

		ignoreFlag := false
		// ignore items to connect
		for _, ignoreAddrItem := range config.Parameters.Base.IgnoreConnectDNSAddrs {
			if v.WalletAddr.ToBase58() == ignoreAddrItem {
				ignoreFlag = true
			}
		}
		if v.WalletAddr.ToBase58() == this.Account.Address.ToBase58() {
			ignoreFlag = true
		}
		if _, err := this.GetDnsPeerPoolItem(hex.EncodeToString(keypair.SerializePublicKey(this.Account.PublicKey))); err != nil {
			ignoreFlag = true
		}
		if ignoreFlag == false {
			err = setDNSNodeFunc(dnsUrl, v.WalletAddr.ToBase58())
			if err != nil {
				continue
			}
		}
	}
	return nil
}
