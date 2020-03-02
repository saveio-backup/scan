package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/saveio/carrier/network"
	"github.com/saveio/dsp-go-sdk/dsp"
	"github.com/saveio/edge/common"
	"github.com/saveio/pylons"
	"github.com/saveio/scan/cmd"
	"github.com/saveio/scan/cmd/flags"
	"github.com/saveio/scan/common/config"
	"github.com/saveio/scan/http/jsonrpc"
	"github.com/saveio/scan/http/localrpc"
	"github.com/saveio/scan/http/restful"
	"github.com/saveio/scan/service"
	"github.com/saveio/scan/storage"
	"github.com/saveio/themis/account"
	thmsCmd "github.com/saveio/themis/cmd"
	thmsComm "github.com/saveio/themis/cmd/common"
	thmsUtils "github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/log"
	thms "github.com/saveio/themis/start"

	"github.com/urfave/cli"
)

func initAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "Save Scan Project."
	app.Action = start
	app.Version = config.VERSION
	app.Copyright = "Copyright in 2019 The Save Authors"
	app.Commands = []cli.Command{
		thmsCmd.AccountCommand,
		thmsCmd.InfoCommand,
		thmsCmd.AssetCommand,
		// thmsCmd.ContractCommand,
		thmsCmd.ImportCommand,
		thmsCmd.ExportCommand,
		thmsCmd.TxCommond,
		thmsCmd.SigTxCommand,
		// thmsCmd.MultiSigAddrCommand,
		// thmsCmd.MultiSigTxCommand,
		thmsCmd.SendTxCommand,
		thmsCmd.ShowTxCommand,
		cmd.ChannelCommand,
		cmd.TrackerCommand,
		cmd.DNSCommand,
	}
	app.Flags = []cli.Flag{
		flags.ScanConfigFlag,

		//common setting
		flags.LogStderrFlag,
		flags.LogLevelFlag,
		flags.RpcServerFlag,
		flags.ConfPathFlag,
		flags.DbDirFlag,
		//p2p setting
		flags.NetworkIDFlag,
		flags.PortFlag,
		flags.SeedListFlag,
		//tracker command setting
		flags.TrackerServerPortFlag,
		flags.TrackerFee,
		flags.WalletFlag,
		//flags.WalletSetFlag,
		//flags.WalletFileFlag,
		flags.HostFlag,
		//ddns command setting
		flags.DnsIpFlag,
		flags.DnsPortFlag,
		flags.DnsAllFlag,
		//channel command setting
		flags.PartnerAddressFlag,
		flags.TargetAddressFlag,
		flags.TotalDepositFlag,
		flags.AmountFlag,
		flags.PaymentIDFlag,
		// RPC settings
		//flags.RPCDisabledFlag,
		//flags.RPCPortFlag,
		//flags.RPCLocalEnableFlag,
		//flags.RPCLocalProtFlag,
		//Restful setting
		//flags.RestfulEnableFlag,
		//flags.RestfulPortFlag,

		//common setting
		thmsUtils.ConfigFlag,
		thmsUtils.DisableEventLogFlag,
		thmsUtils.DataDirFlag,
		//account setting
		thmsUtils.WalletFileFlag,
		thmsUtils.AccountAddressFlag,
		thmsUtils.AccountPassFlag,
		//consensus setting
		thmsUtils.EnableConsensusFlag,
		thmsUtils.MaxTxInBlockFlag,
		//txpool setting
		thmsUtils.GasPriceFlag,
		thmsUtils.GasLimitFlag,
		thmsUtils.TxpoolPreExecDisableFlag,
		thmsUtils.DisableSyncVerifyTxFlag,
		thmsUtils.DisableBroadcastNetTxFlag,
		//p2p setting
		thmsUtils.ReservedPeersOnlyFlag,
		thmsUtils.ReservedPeersFileFlag,
		thmsUtils.NetworkIdFlag,
		thmsUtils.NodePortFlag,
		thmsUtils.ConsensusPortFlag,
		thmsUtils.DualPortSupportFlag,
		thmsUtils.MaxConnInBoundFlag,
		thmsUtils.MaxConnOutBoundFlag,
		thmsUtils.MaxConnInBoundForSingleIPFlag,
		//test mode setting
		thmsUtils.EnableTestModeFlag,
		thmsUtils.TestModeGenBlockTimeFlag,
		//rpc setting
		thmsUtils.RPCDisabledFlag,
		thmsUtils.RPCPortFlag,
		thmsUtils.RPCLocalEnableFlag,
		thmsUtils.RPCLocalProtFlag,
		//rest setting
		thmsUtils.RestfulEnableFlag,
		thmsUtils.RestfulPortFlag,
		//ws setting
		// thmsUtils.WsEnabledFlag,
		thmsUtils.WsPortFlag,
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
	if err := initAPP().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func start(ctx *cli.Context) {
	cmd.InitLog(ctx)

	acc, err := getDefaultAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if !config.Parameters.Base.DisableChain {
		logwithsideline("CHAIN STARTING")
		startChain(ctx)
	}
	logwithsideline("SCAN STARTING")
	startScan(ctx, acc)
	logwithsideline("SETUP DOWN, ENJOY!!!")
	if config.Parameters.Base.DumpMemory == true {
		go dumpMemory()
	}
	WaitToExit()
}

func startChain(ctx *cli.Context) {
	_, err := thms.InitConfig(ctx)
	if err != nil {
		log.Errorf("startChain initConfig error:%s", err)
		return
	}

	acc, err := thms.InitAccount(ctx)
	if err != nil {
		log.Errorf("startChain InitAccount error:%s", err)
		return
	}
	_, err = thms.InitLedger(ctx)
	if err != nil {
		log.Errorf("startChain initLedger error:%s", err)
		return
	}
	//defer ldg.Close()
	txpool, err := thms.InitTxPool(ctx)
	if err != nil {
		log.Errorf("startChain initTxPool error:%s", err)
		return
	}
	p2pSvr, p2pPid, err := thms.InitP2PNode(ctx, txpool)
	if err != nil {
		log.Errorf("startChain initP2PNode error:%s", err)
		return
	}
	_, err = thms.InitConsensus(ctx, p2pPid, txpool, acc)
	if err != nil {
		log.Errorf("startChain initConsensus error:%s", err)
		return
	}
	err = thms.InitRpc(ctx)
	if err != nil {
		log.Errorf("startChain initRpc error:%s", err)
		return
	}
	err = thms.InitLocalRpc(ctx)
	if err != nil {
		log.Errorf("startChain initLocalRpc error:%s", err)
		return
	}
	thms.InitRestful(ctx)
	thms.InitWs(ctx)
	thms.InitNodeInfo(ctx, p2pSvr)

	go thms.LogCurrBlockHeight()
	log.Info("Chain started SUCCESS")
}

func startScan(ctx *cli.Context, acc *account.Account) {
	config.Init(ctx)
	service.Init(acc)

	edb, err := storage.NewLevelDBStore(config.EndpointDBPath())
	if err != nil {
		log.Fatal(err)
	}
	storage.EDB = storage.NewEndpointDB(edb)
	tdb, err := storage.NewLevelDBStore(config.TrackerDBPath())
	if err != nil {
		log.Fatal(err)
	}
	storage.TDB = storage.NewTorrentDB(tdb)

	startChannelNetwork, startDnsNetwork, startTkNetwork := true, true, true
	err = service.ScanNode.StartScanNode(startChannelNetwork, startDnsNetwork, startTkNetwork)
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
	// err = service.ScanNode.StartTrackerService()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	if err := initRpc(ctx); err != nil {
		log.Fatalf("rpc start err: %v", err.Error())
	}

	log.Info("rpc start success.")
	initRestful(ctx)
	log.Info("restful start success.")

	err = service.ScanNode.StartChannelService()
	if err != nil {
		log.Fatalf("channel service start err : %v", err)
	}
	log.Info("channel service started.")
	log.Infof("scan version: %s", config.VERSION)
	log.Infof("dsp-go-sdk version: %s", dsp.Version)
	log.Infof("pylons version: %s", pylons.Version)
	log.Infof("carrier version: %s", network.Version)
}

func getDefaultAccount(ctx *cli.Context) (*account.Account, error) {
	walletFile := ctx.GlobalString(thmsUtils.GetFlagName(thmsUtils.WalletFileFlag))
	if walletFile == "" {
		return nil, fmt.Errorf("Please config wallet file using --wallet flag")
	}
	if !common.FileExisted(walletFile) {
		return nil, fmt.Errorf("Cannot find wallet file:%s. Please create wallet first", walletFile)
	}

	acc, err := thmsComm.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account error:%s", err)
	}
	fmt.Printf("\nUsing account: %s\n", acc.Address.ToBase58())
	return acc, nil
}

func initRpc(ctx *cli.Context) error {
	if !config.Parameters.Base.EnableJsonRpc {
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
	return nil
}

func initLocalRpc(ctx *cli.Context) error {
	log.Infof("localRpc start to init...")
	if !ctx.GlobalBool(flags.GetFlagName(flags.RPCLocalEnableFlag)) {
		return nil
	}
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
	return nil
}

func initRestful(ctx *cli.Context) {
	if !config.Parameters.Base.EnableRest {
		return
	}
	go restful.StartServer()
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

func WaitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("seeds received exit signal:%v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}

func dumpMemory() {
	i := 1
	for {
		filename := fmt.Sprintf("Heap.prof.%d", i)

		f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Tracef("Heap Profile generated FAILED")
			log.Fatal(err)
			break
		}

		log.Info("Heap Profile %s generated", filename)

		time.Sleep(30 * time.Second)
		pprof.WriteHeapProfile(f)
		f.Close()
		i++
	}
}
