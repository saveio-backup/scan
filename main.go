package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/saveio/scan/channel"
	"github.com/saveio/scan/dns"
	"github.com/saveio/scan/tracker"

	chaincmd "github.com/saveio/themis/cmd"
	chainutils "github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/log"

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
	"github.com/saveio/scan/network/actor/recv"
	"github.com/saveio/scan/storage"
	tcomm "github.com/saveio/scan/tracker/common"
	"github.com/urfave/cli"

	//"github.com/saveio/themis-go-sdk/wallet"
	//"github.com/saveio/themis-go-sdk/wallet"
	//"github.com/saveio/scan/tracker"
	"github.com/saveio/themis/errors"
)

func initAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "Save DDNS"
	app.Action = startDDNS
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

func startDDNS(ctx *cli.Context) {
	config.DefaultConfig = config.GenDefConfig()
	log.Info("\n------------------\n/> Chain starting \n------------------")
	//chainnode
	err := startChain(ctx)
	if err != nil {
		log.Error("Chain init error")
		return
	}
	log.Info("Chain start success")
	log.Info("\n----------------------------------------\n/> Dns Tracker Channel Service starting \n----------------------------------------")
	startTracker(ctx)
	common.WaitToExit()
}

func startTracker(ctx *cli.Context) {
	initLog(ctx)
	cmd.SetDDNSConfig(ctx)
	p2pSvr, p2pPid, err := initP2P(ctx)
	if err != nil {
		log.Errorf("initP2PNode error:%s", err)
		return
	}

	svr := netserver.NewNetServer()
	svr.Tsvr.SetPID(p2pPid)
	if err = svr.Run(); err != nil {
		log.Errorf("run ddns server error:%s", err)
		return
	}

	network.DDNSP2P = p2pSvr
	storage.TDB, err = initTrackerDB(config.DefaultConfig.CommonConfig.CommonDBPath)
	if err != nil {
		log.Fatalf("InitTrackerDB error: %v", err)
	}

	err = initRpc(ctx)
	if err != nil {
		log.Errorf("initTrackerRpc error:%s", err)
		return
	}
	initRestful(ctx)
	initChannelService(ctx, p2pSvr.GetPID())
	initDnsService(ctx)
	// initEndPointReg(ctx, p2pSvr)
	log.Info("Tracker start success")
	log.Info("\n--------------\n/> Setup Done \n--------------")
}

func initLog(ctx *cli.Context) {
	//init log module
	cmd.SetLogConfig(ctx)
	log.Info("start logging...")
}

func initTrackerDB(path string) (*storage.LevelDBStore, error) {
	log.Info("Tracker DB is init...")
	db, err := storage.NewLevelDBStore(path)
	if err != nil {
		return nil, err
	}
	log.Info("Tracker DB init success")
	return db, nil
}

func initP2P(ctx *cli.Context) (*network.Network, *actor.PID, error) {
	log.Info("P2P is init...")
	// wallet, err := ccom.OpenWallet(ctx)
	// if err != nil {
	// 	return nil, nil, err
	// }
	// pwd := []byte(config.DefaultConfig.CommonConfig.WalletPwd)

	// account, err := wallet.GetDefaultAccount(pwd)
	// if err != nil {
	// 	log.Errorf("initP2P GetDefaultAccount error:%s\n", err)
	// 	return nil, nil, err
	// }

	p2pserver := network.NewP2P()
	// bPrivate := keypair.SerializePrivateKey(account.PrivKey())
	// bPub := keypair.SerializePublicKey(account.PubKey())
	// p2pserver.Keys = &crypto.KeyPair{
	// 	PrivateKey: bPrivate,
	// 	PublicKey:  bPub,
	// }
	go p2pserver.Start(fmt.Sprintf("tcp://127.0.0.1:%d", config.DefaultConfig.CommonConfig.P2PNATPort))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	pid, err := recv.NewP2PActor(p2pserver)
	if err != nil {
		log.Fatal(err)
	}

	// p2p := network.NewP2P()
	// go p2p.Start()
	// pid, err := recv.NewP2PActor(p2p)
	// BlockUntilComplete()
	if err != nil {
		return nil, nil, fmt.Errorf("p2pActor init error %s", err)
	}
	p2pserver.SetPID(pid)
	log.Infof("PID:%v", pid)
	log.Info("p2p init success")
	return p2pserver, pid, nil
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

func startChain(ctx *cli.Context) error {
	chain.InitLog(ctx)

	_, err := chain.InitConfig(ctx)
	if err != nil {
		log.Errorf("initConfig error:%s", err)
		return err
	}
	acc, err := chain.InitAccount(ctx)
	if err != nil {
		log.Errorf("initWallet error:%s", err)
		return err
	}
	_, err = chain.InitLedger(ctx)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}
	//defer ldg.Close()
	txpool, err := chain.InitTxPool(ctx)
	if err != nil {
		log.Errorf("initTxPool error:%s", err)
		return err
	}
	p2pSvr, p2pPid, err := chain.InitP2PNode(ctx, txpool)
	if err != nil {
		log.Errorf("initP2PNode error:%s", err)
		return err
	}
	_, err = chain.InitConsensus(ctx, p2pPid, txpool, acc)
	if err != nil {
		log.Errorf("initConsensus error:%s", err)
		return err
	}
	err = chain.InitRpc(ctx)
	if err != nil {
		log.Errorf("initRpc error:%s", err)
		return err
	}
	err = chain.InitLocalRpc(ctx)
	if err != nil {
		log.Errorf("initLocalRpc error:%s", err)
		return err
	}
	chain.InitRestful(ctx)
	chain.InitWs(ctx)
	chain.InitNodeInfo(ctx, p2pSvr)

	go chain.LogCurrBlockHeight()
	return nil
}

func initEndPointReg(ctx *cli.Context, p2p *network.Network) {
	client, err := ccom.OpenWallet(ctx)
	if err != nil {
		log.Error("initEndPointReg OpenWallet error")
	}
	// pw, err := ccom.GetPasswd(ctx)
	// if err != nil {
	// 	log.Errorf("regEndPoint GetPasswd error:%s\n", err)
	// 	return err
	// }
	pw := []byte(config.DefaultConfig.CommonConfig.WalletPwd)
	acc, err := client.GetDefaultAccount(pw)
	if err != nil {
		log.Errorf("regEndPoint GetDefaultAccount error:%s\n", err)
	}
	a := acc.Address
	if ctx.IsSet(utils.GetFlagName(utils.WalletFlag)) {
		wAddr = ctx.String(utils.GetFlagName(utils.WalletFlag))
	} else {
		wAddr = a.ToBase58()
	}
	log.Infof("ExternalAddr is %s\n", p2p.ListenAddr())
	if p2p.ListenAddr() != "" {
		host := p2p.ListenAddr()
		err = tracker.EndPointRegistry(wAddr, host)
		if err != nil {
			log.Errorf("DDNS node EndPointRegistry error:%v", err)
		}
	} else {
		log.Error("DDNS node can not acquire external ip")
	}
	log.Info("initEndPointReg success!")
}

func initChannelService(ctx *cli.Context, p2pActor *actor.PID) error {
	wallet, err := ccom.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pwd := []byte(config.DefaultConfig.CommonConfig.WalletPwd)

	acc, err := wallet.GetDefaultAccount(pwd)
	if err != nil {
		log.Errorf("initChannelService GetDefaultAccount error:%s\n", err)
		return err
	}
	channel.GlbChannelSvr, err = channel.NewChannelSvr(acc, p2pActor)
	if err != nil {
		return errors.NewErr("initChannelService Failed.")
	}
	err = channel.GlbChannelSvr.Start()
	if err != nil {
		return err
	}
	log.Info("initChannelService success!")
	return nil
}

func initDnsService(ctx *cli.Context) error {
	wallet, err := ccom.OpenWallet(ctx)
	if err != nil {
		return err
	}
	// pwd, err := ccom.GetPasswd(ctx)
	// if err != nil {
	// 	log.Errorf("initGlobalDspService GetPasswd error:%s\n", err)
	// 	return err
	// }
	pwd := []byte(config.DefaultConfig.CommonConfig.WalletPwd)

	acc, err := wallet.GetDefaultAccount(pwd)
	if err != nil {
		log.Errorf("initDnsService GetDefaultAccount error:%s\n", err)
		return err
	}
	dns.GlbDNSSvr, err = dns.NewDNSSvr(acc)
	if err != nil {
		return errors.NewErr("initDnsService Failed.")
	}
	log.Info("initDnsService success!")
	return nil
}
