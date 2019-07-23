package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/saveio/dsp-go-sdk/actor/client"
	"github.com/saveio/scan/channel"
	"github.com/saveio/scan/dns"
	"github.com/saveio/scan/network/actor/server"
	"github.com/saveio/scan/tracker"

	chaincmd "github.com/saveio/themis/cmd"
	chainutils "github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/crypto/keypair"

	"github.com/saveio/scan/cmd"
	chain "github.com/saveio/themis/start"

	//ccom "github.com/saveio/scan/cmd/common"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	ccom "github.com/saveio/scan/cmd/common"
	"github.com/saveio/scan/cmd/utils"
	"github.com/saveio/scan/common"
	"github.com/saveio/scan/common/config"
	"github.com/saveio/scan/http/jsonrpc"
	"github.com/saveio/scan/http/localrpc"
	"github.com/saveio/scan/http/restful"
	"github.com/saveio/scan/netserver"
	"github.com/saveio/scan/network"
	"github.com/saveio/scan/storage"
	tcomm "github.com/saveio/scan/tracker/common"
	"github.com/urfave/cli"

	//"github.com/saveio/themis-go-sdk/wallet"
	//"github.com/saveio/themis-go-sdk/wallet"
	//"github.com/saveio/scan/tracker"
	"github.com/saveio/themis/account"
	cutils "github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/errors"
)

func initAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "Save DDNS"
	app.Action = start
	app.Version = config.VERSION
	app.Copyright = "Copyright in 2019 The Save Authors"
	app.Commands = []cli.Command{
		chaincmd.AccountCommand,
		chaincmd.InfoCommand,
		chaincmd.AssetCommand,
		chaincmd.ContractCommand,
		chaincmd.ImportCommand,
		chaincmd.ExportCommand,
		chaincmd.TxCommond,
		chaincmd.SigTxCommand,
		chaincmd.MultiSigAddrCommand,
		chaincmd.MultiSigTxCommand,
		chaincmd.SendTxCommand,
		chaincmd.ShowTxCommand,
		cmd.EndPointCommand,
		cmd.ChannelCommand,
		cmd.DNSCommand,
	}
	app.Flags = []cli.Flag{
		//common setting
		utils.LogStderrFlag,
		utils.LogLevelFlag,
		utils.RpcServerFlag,
		utils.ConfPathFlag,
		utils.DbDirFlag,
		//p2p setting
		utils.NetworkIDFlag,
		utils.PortFlag,
		utils.SeedListFlag,
		//tracker command setting
		utils.TrackerServerPortFlag,
		utils.TrackerFee,
		utils.WalletFlag,
		//utils.WalletSetFlag,
		//utils.WalletFileFlag,
		utils.HostFlag,
		//ddns command setting
		utils.DnsIpFlag,
		utils.DnsPortFlag,
		utils.DnsAllFlag,
		//channel command setting
		utils.PartnerAddressFlag,
		utils.TargetAddressFlag,
		utils.TotalDepositFlag,
		utils.AmountFlag,
		utils.PaymentIDFlag,
		// RPC settings
		//utils.RPCDisabledFlag,
		//utils.RPCPortFlag,
		//utils.RPCLocalEnableFlag,
		//utils.RPCLocalProtFlag,
		//Restful setting
		//utils.RestfulEnableFlag,
		//utils.RestfulPortFlag,
		//common setting
		chainutils.ConfigFlag,
		chainutils.DisableEventLogFlag,
		chainutils.DataDirFlag,
		//account setting
		chainutils.WalletFileFlag,
		chainutils.AccountAddressFlag,
		chainutils.AccountPassFlag,
		//consensus setting
		chainutils.EnableConsensusFlag,
		chainutils.MaxTxInBlockFlag,
		//txpool setting
		chainutils.GasPriceFlag,
		chainutils.GasLimitFlag,
		chainutils.TxpoolPreExecDisableFlag,
		chainutils.DisableSyncVerifyTxFlag,
		chainutils.DisableBroadcastNetTxFlag,
		//p2p setting
		chainutils.ReservedPeersOnlyFlag,
		chainutils.ReservedPeersFileFlag,
		chainutils.NetworkIdFlag,
		chainutils.NodePortFlag,
		chainutils.ConsensusPortFlag,
		chainutils.DualPortSupportFlag,
		chainutils.MaxConnInBoundFlag,
		chainutils.MaxConnOutBoundFlag,
		chainutils.MaxConnInBoundForSingleIPFlag,
		//test mode setting
		chainutils.EnableTestModeFlag,
		chainutils.TestModeGenBlockTimeFlag,
		//rpc setting
		chainutils.RPCDisabledFlag,
		chainutils.RPCPortFlag,
		chainutils.RPCLocalEnableFlag,
		chainutils.RPCLocalProtFlag,
		//rest setting
		chainutils.RestfulEnableFlag,
		chainutils.RestfulPortFlag,
		//ws setting
		// chainutils.WsEnabledFlag,
		chainutils.WsPortFlag,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		cmd.Init(context)
		return nil
	}
	return app
}

var wAddr string

func main() {
	fmt.Print(os.Args)
	if err := initAPP().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func logwithsideline(msg string) {
	var line = ""
	for len(line) < len(msg)+4 {
		line += "-"
	}
	fmt.Println(line)
	fmt.Printf("/> %s\n", msg)
	fmt.Println(line)
}

func start(ctx *cli.Context) {
	fmt.Print("\n")
	config.SetupDefaultConfig()
	logwithsideline("CHAIN STARTING")
	startChain(ctx)
	logwithsideline("SCAN INITIALIZE STARTING")
	initialize(ctx)
	logwithsideline("SETUP DOWN, ENJOY!!!")
	common.WaitToExit()
}

func startChain(ctx *cli.Context) {
	// chain.InitLog(ctx)

	_, err := chain.InitConfig(ctx)
	if err != nil {
		log.Errorf("startChain initConfig error:%s", err)
		return
	}
	acc, err := chain.InitAccount(ctx)
	if err != nil {
		log.Errorf("startChain initWallet error:%s", err)
		return
	}
	_, err = chain.InitLedger(ctx)
	if err != nil {
		log.Errorf("startChain initLedger error:%s", err)
		return
	}
	//defer ldg.Close()
	txpool, err := chain.InitTxPool(ctx)
	if err != nil {
		log.Errorf("startChain initTxPool error:%s", err)
		return
	}
	p2pSvr, p2pPid, err := chain.InitP2PNode(ctx, txpool)
	if err != nil {
		log.Errorf("startChain initP2PNode error:%s", err)
		return
	}
	_, err = chain.InitConsensus(ctx, p2pPid, txpool, acc)
	if err != nil {
		log.Errorf("startChain initConsensus error:%s", err)
		return
	}
	err = chain.InitRpc(ctx)
	if err != nil {
		log.Errorf("startChain initRpc error:%s", err)
		return
	}
	err = chain.InitLocalRpc(ctx)
	if err != nil {
		log.Errorf("startChain initLocalRpc error:%s", err)
		return
	}
	chain.InitRestful(ctx)
	chain.InitWs(ctx)
	chain.InitNodeInfo(ctx, p2pSvr)

	go chain.LogCurrBlockHeight()
	log.Info("Chain started SUCCESS")
}

func initialize(ctx *cli.Context) {
	//init log module
	cmd.SetLogConfig(ctx)
	log.Info("start logging...")

	// get account
	acc, balance, err := getDefaultAccount(ctx)
	if err != nil {
		log.Errorf("SCAN initialize getDefaultAccount FAILED, err: %v", err)
		// os.Exit(1)
	}

	// setup actor
	p2pActor, err := server.NewP2PActor()
	if err != nil {
		log.Errorf("SCAN initialize NewP2pActor FAILED, err: %v", err)
		os.Exit(1)
	}
	log.Info("SCAN initialize NewP2pActor SUCCESS.")

	// setup tracker service
	storage.TDB, err = storage.NewLevelDBStore(config.DefaultConfig.CommonConfig.CommonDBPath)
	if err != nil {
		log.Errorf("SCAN Initialize TrackerDB FAILED, err: %v", err)
		os.Exit(1)
	}
	log.Info("SCAN initialize TrackerDB SUCCESS.")

	// setup channel
	channel.GlbChannelSvr, err = channel.NewChannelSvr(acc, p2pActor.GetLocalPID())
	if err != nil {
		log.Errorf("SCAN initialize NewChannelSvr FAILED, err: %v", err)
		os.Exit(1)
	}
	log.Info("SCAN initialize NewChannelSvr SUCCESS.")

	// setup dns
	dns.GlbDNSSvr, err = dns.NewDNSSvr(acc)
	if err != nil {
		log.Errorf("SCAN initialize NewDNSSvr FAILED, err: %v", err)
		os.Exit(1)
	}
	log.Info("SCAN initialize NewDNSSvr SUCCESS.")

	// setup p2p network
	p2pNetwork := network.NewP2P()
	p2pNetwork.SetProxyServer(config.DefaultConfig.CommonConfig.P2PNATAddr)
	p2pNetwork.SetPID(p2pActor.GetLocalPID())
	p2pActor.SetNetwork(p2pNetwork)
	network.DDNSP2P = p2pNetwork

	// fetch all dns nodes for p2p bootstraps
	var bootstraps []string
	ns, err := dns.GlbDNSSvr.GetAllDnsNodes()
	if err != nil {
		log.Error(err)
	} else {
		for _, v := range ns {
			ignoreFlag := false
			for _, ignoreAddrItem := range config.DefaultConfig.DnsConfig.IgnoreConnectDNSAddrs {
				if v.WalletAddr.ToBase58() == ignoreAddrItem {
					ignoreFlag = true
				}
			}
			if _, err := dns.GlbDNSSvr.GetDnsPeerPoolItem(hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey))); err != nil {
				ignoreFlag = true
			}

			if !ignoreFlag {
				bootstraps = append(bootstraps, fmt.Sprintf("%s://%s:%s", config.DefaultConfig.P2PConfig.Protocol, string(v.IP), string(v.Port)))
			}
		}
	}

	if err := p2pNetwork.Start(common.FullHostAddr(
		fmt.Sprintf("127.0.0.1:%d", config.DefaultConfig.P2PConfig.PortBase),
		config.DefaultConfig.P2PConfig.Protocol), bootstraps); err != nil {
		log.Errorf("SCAN initialize Start P2P FAILED, err: %v", err)
		os.Exit(1)
	}
	log.Info("SCAN initialize Start P2P SUCCESS.")

	// setup dns and endpoint registry
	if p2pNetwork.PublicAddr() != "" {
		log.Infof("SCAN initialize External ListenAddr is %s", p2pNetwork.PublicAddr())
		// start tracker server
		netSvr := netserver.NewNetServer()
		netSvr.Tsvr.SetPID(p2pActor.GetLocalPID())
		if err = netSvr.Run(); err != nil {
			log.Errorf("SCAN Initialize TrackerServer Run FAILED, err: %v", err)
			os.Exit(1)
		}
		log.Info("SCAN initialize TrackerServer Run SUCCESS.")

		if config.DefaultConfig.DnsConfig.AutoSetupDNSRegisterEnable {
			autoSetupDNSRegisterWorking(ctx, acc, p2pNetwork.PublicAddr(), balance)
		}
	} else {
		log.Error("SCAN initialize GET PublicAddr FAILED, can not acquire external ip")
		os.Exit(1)
	}

	// start channel service
	if err = channel.GlbChannelSvr.StartChannelService(); err != nil {
		log.Errorf("SCAN initialize StartChannelService FAILED, err: %v", err)
		os.Exit(1)
	}
	log.Info("SCAN initialize StartChannelService SUCCESS.")

	// setup rpc & restful server
	if err := initRpc(ctx); err != nil {
		log.Errorf("SCAN initialize initRpc FAILED, err: %v", err.Error())
		os.Exit(1)
	}
	initRestful(ctx)

	// setup dns channel connect
	if config.DefaultConfig.DnsConfig.AutoSetupDNSChannelsEnable {
		go autoSetupDNSChannelsWorking(ctx, p2pActor.GetLocalPID())
	}
}

func getDefaultAccount(ctx *cli.Context) (*account.Account, uint64, error) {
	wallet, err := ccom.OpenWallet(ctx)
	if err != nil {
		return nil, 0, err
	}
	pwd := []byte(config.DefaultConfig.CommonConfig.WalletPwd)

	acc, err := wallet.GetDefaultAccount(pwd)
	if err != nil {
		return nil, 0, err
	}

	// Check Balance
	bal, err := cutils.GetBalance(acc.Address.ToBase58())
	if err != nil {
		return nil, 0, err
	}
	balance, err := strconv.ParseUint(bal.Usdt, 10, 64)
	if err := cutils.CheckAssetAmount("usdt", balance); err != nil {
		log.Errorf("BalanceOf default wallet address %s CheckAssetAmount err: %v, DNS NODE CAN NOT WORK.", acc.Address.ToBase58(), err)
		return acc, 0, err
	} else if balance <= 0 {
		log.Errorf("BalanceOf default wallet address %s is: %d, DNS NODE CAN NOT WORK.", acc.Address.ToBase58(), balance)
		return acc, 0, err
	}

	return acc, balance, nil
}

func initRpc(ctx *cli.Context) error {
	if !config.DefaultConfig.RpcConfig.EnableHttpJsonRpc {
		return nil
	}
	var err error
	exitCh := make(chan interface{}, 0)
	go func() {
		err = jsonrpc.StartRPCServer()
		close(exitCh)
	}()

	flag := false
	select {
	case <-exitCh:
		if !flag {
			return err
		}
	case <-time.After(time.Millisecond * 5):
		flag = true
	}
	log.Infof("Rpc init success jsonrpc")
	return nil
}

func initLocalRpc(ctx *cli.Context) error {
	log.Infof("localRpc start to init...")
	//if !ctx.GlobalBool(utils.GetFlagName(utils.RPCLocalEnableFlag)) {
	//	return nil
	//}
	var err error
	exitCh := make(chan interface{}, 0)
	go func() {
		err = localrpc.StartLocalServer()
		close(exitCh)
	}()

	flag := false
	select {
	case <-exitCh:
		if !flag {
			return err
		}
	case <-time.After(time.Millisecond * 5):
		flag = true
	}

	log.Infof("Local rpc init success")
	return nil
}

func initRestful(ctx *cli.Context) {
	if !config.DefaultConfig.RestfulConfig.EnableHttpRestful {
		return
	}
	go restful.StartServer()

	log.Infof("Restful init success")
}

func BlockUntilComplete() {
	<-tcomm.ListeningCh
}

func autoSetupDNSRegisterWorking(ctx *cli.Context, acc *account.Account, p2pPublicAddr string, balance uint64) error {
	publicAddr, err := common.SplitHostAddr(p2pPublicAddr)
	if err != nil {
		log.Fatal("SCAN initialize External ListenAddr Split FAILED, err", err)
		os.Exit(1)
	}

	dnsinfo, err := dns.GlbDNSSvr.GetDnsNodeByAddr(acc.Address)
	log.Infof("SCAN initialize DNS getbyaddr dnsinfo: %v, err: %v", dnsinfo, err)
	if dnsinfo != nil {
		if _, err := dns.GlbDNSSvr.DNSNodeUpdate(publicAddr.Host, strings.Split(p2pPublicAddr, ":")[2]); err != nil {
			log.Fatalf("SCAN initialize DNS update FAILED, err: %v", err)
			os.Exit(1)
		} else {
			log.Infof("SCAN initialize DNS update hostinfo SUCCESS.")
		}
	} else {
		if _, err = dns.GlbDNSSvr.DNSNodeReg(publicAddr.Host, publicAddr.Port,
			config.DefaultConfig.DnsConfig.InitDeposit); err != nil {
			log.Infof("BalaceOf Default Wallet Address %s is: %s, DNS InitDeposit needs: %s",
				acc.Address.ToBase58(),
				cutils.FormatUsdt(balance),
				cutils.FormatUsdt(config.DefaultConfig.DnsConfig.InitDeposit))
			log.Fatalf("SCAN initialize DNS register FAILED, err: %v", err)
			os.Exit(1)
		} else {
			log.Info("SCAN initialize DNS register SUCCESS.")
		}
	}
	if err = tracker.EndPointRegistry(acc.Address.ToBase58(),
		fmt.Sprintf("%s:%s", publicAddr.Host, publicAddr.Port)); err != nil {
		log.Errorf("SCAN initialize EndPointRegistry FAILED, err:%v", err)
		os.Exit(1)
	} else {
		log.Info("SCAN initialize EndPointRegistry SUCCESS.")
	}
	return nil
}

func autoSetupDNSChannelsWorking(ctx *cli.Context, p2pActor *actor.PID) error {
	wallet, err := ccom.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pwd := []byte(config.DefaultConfig.CommonConfig.WalletPwd)

	acc, err := wallet.GetDefaultAccount(pwd)
	if err != nil {
		log.Errorf("autoSetupDNSChannelsWorking GetDefaultAccount error:%s\n", err)
		return err
	}

	client.SetP2pPid(p2pActor)

	ns, err := dns.GlbDNSSvr.GetAllDnsNodes()
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
		err = channel.GlbChannelSvr.Channel.SetHostAddr(walletAddr, dnsUrl)
		if err != nil {
			return err
		}
		_, err = channel.GlbChannelSvr.Channel.OpenChannel(walletAddr, 0)
		if err != nil {
			log.Debugf("open channel err ")
			return err
		}
		err = channel.GlbChannelSvr.Channel.WaitForConnected(walletAddr, time.Duration(100)*time.Second)
		if err != nil {
			log.Errorf("wait channel connected err %s %s", walletAddr, err)
			return err
		}
		log.Debugf("channel connected %s %s", walletAddr, err)
		bal, _ := channel.GlbChannelSvr.Channel.GetAvailableBalance(walletAddr)
		log.Debugf("current balance %d", bal)
		log.Infof("connect to dns node :%s, deposit %d", dnsUrl, config.DefaultConfig.DnsConfig.ChannelDeposit)
		err = channel.GlbChannelSvr.Channel.SetDeposit(walletAddr, config.DefaultConfig.DnsConfig.ChannelDeposit)
		if err != nil && strings.Index(err.Error(), "totalDeposit must big than contractBalance") == -1 {
			log.Debugf("deposit result %s", err)
			// TODO: withdraw and close channel
			return err
		}
		bal, _ = channel.GlbChannelSvr.Channel.GetAvailableBalance(walletAddr)
		log.Debugf("current deposited balance %d", bal)
		log.Info("channel deposit success")

		err := channel.GlbChannelSvr.Channel.CanTransfer(walletAddr, 10)
		if err == nil {
			log.Info("loopTest can transfer!")
			for i := 0; i < 10; i++ {
				err := channel.GlbChannelSvr.Transfer(1, 1, walletAddr)
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
		log.Debugf("autoSetupDNSChannelsWorking range DNS %s :%v, port %v, ", v.WalletAddr.ToBase58(), string(v.IP), string(v.Port))

		// dnsUrl, _ := GetExternalIP(v.WalletAddr.ToBase58())
		var dnsUrl string
		if dnsInfo, err := dns.GlbDNSSvr.GetDnsNodeByAddr(v.WalletAddr); dnsInfo != nil {
			dnsUrl = common.FullHostAddr(fmt.Sprintf("%s:%s", dnsInfo.IP, dnsInfo.Port), config.DefaultConfig.DnsConfig.Protocol)
		} else {
			log.Errorf("autoSetupDNSChannelsWorking GetDnsNodeByAddr Failed. err:%v", err)
		}

		if len(dnsUrl) == 0 {
			dnsUrl = fmt.Sprintf("%s://%s:%s", config.DefaultConfig.DnsConfig.Protocol, v.IP, v.Port)
		}

		if err = tracker.EndPointRegistry(v.WalletAddr.ToBase58(), strings.Split(dnsUrl, "://")[1]); err != nil {
			log.Errorf("autoSetupDNSChannelsWorking tracker.EndPointRegistry Failed. err:%v", err)
		}

		ignoreFlag := false
		// ignore items to connect
		for _, ignoreAddrItem := range config.DefaultConfig.DnsConfig.IgnoreConnectDNSAddrs {
			if v.WalletAddr.ToBase58() == ignoreAddrItem {
				ignoreFlag = true
			}
		}

		if v.WalletAddr.ToBase58() == acc.Address.ToBase58() {
			ignoreFlag = true
		}

		if _, err := dns.GlbDNSSvr.GetDnsPeerPoolItem(hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey))); err != nil {
			ignoreFlag = true
		}

		if ignoreFlag == false {
			err = setDNSNodeFunc(dnsUrl, v.WalletAddr.ToBase58())
			if err != nil {
				continue
			}
		}
	}
	return err
}

//GetExternalIP. get external ip of wallet from dns nodes
func GetExternalIP(walletAddr string) (string, error) {
	hostAddr, err := tracker.EndPointQuery(walletAddr)
	if err != nil {
		log.Errorf("address from req failed %s", err)
		return "", err
	}
	log.Debugf("GetExternalIP %s :%v", walletAddr, string(hostAddr))
	if len(string(hostAddr)) == 0 {
		return "", errors.NewErr("host addr not found")
	}
	hostAddrStr := hostAddr
	if strings.Index(hostAddrStr, "0.0.0.0:0") != -1 {
		return "", errors.NewErr("host addr format wrong")
	}
	return hostAddrStr, nil
}
