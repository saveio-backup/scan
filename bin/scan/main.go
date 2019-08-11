package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/saveio/scan/service"

	chaincmd "github.com/saveio/themis/cmd"
	chainutils "github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/log"

	"github.com/saveio/scan/cmd"
	chain "github.com/saveio/themis/start"

	//ccom "github.com/saveio/scan/cmd/common"
	"time"

	"github.com/saveio/scan/cmd/flags"
	"github.com/saveio/scan/common"
	"github.com/saveio/scan/common/config"
	"github.com/saveio/scan/http/jsonrpc"
	"github.com/saveio/scan/http/localrpc"
	"github.com/saveio/scan/http/restful"
	"github.com/saveio/scan/storage"
	"github.com/urfave/cli"
	//"github.com/saveio/themis-go-sdk/wallet"
	//"github.com/saveio/themis-go-sdk/wallet"
	//"github.com/saveio/scan/tracker"
)

func initAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "Save Scan Project."
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

func start(ctx *cli.Context) {
	fmt.Print("\n")
	_, err := chain.InitConfig(ctx)
	if err != nil {
		log.Errorf("startChain initConfig error:%s", err)
		return
	}
	config.Init(ctx)
	config.SetScanConfig(ctx)
	if config.Parameters.Base.DisableChain == false {
		logwithsideline("CHAIN STARTING")
		startChain(ctx)
	}
	logwithsideline("SCAN STARTING")
	startScan(ctx)
	logwithsideline("SETUP DOWN, ENJOY!!!")
	if config.Parameters.Base.DumpMemory == true {
		go dumpMemory()
	}
	common.WaitToExit()
}

func startChain(ctx *cli.Context) {
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

func startScan(ctx *cli.Context) {
	//init log module
	cmd.SetLogConfig(ctx)
	log.Info("start logging...")

	service.Init(config.Parameters.Base.WalletDir, config.Parameters.Base.WalletPwd)

	startChannelNetwork, startDnsNetwork := true, true
	err := service.ScanNode.StartScanNode(startChannelNetwork, startDnsNetwork)
	if err != nil {
		log.Fatal(err)
	}
	storage.TDB, err = storage.NewLevelDBStore(config.TrackerDBPath())
	if err != nil {
		log.Fatal(err)
	}
	err = service.ScanNode.StartTrackerService()
	if err != nil {
		log.Fatal(err)
	}

	epDB, err := storage.NewLevelDBStore(config.EndpointDBPath())
	if err != nil {
		log.Fatal(err)
	}
	storage.EDB = storage.NewEndpointDB(epDB)

	if err := initRpc(ctx); err != nil {
		log.Fatalf("start scan rpc err: %v", err.Error())
	}
	log.Info("start scan rpc success.")
	initRestful(ctx)
	log.Info("start scan restful success.")

	err = service.ScanNode.StartChannelService()
	if err != nil {
		log.Fatalf("start scan start channel service err : %v", err)
	}
	log.Info("start scan start channel service success.")

	// setup dns channel connect
	if config.Parameters.Base.AutoSetupDNSChannelsEnable {
		go service.ScanNode.AutoSetupDNSChannelsWorking()
	}
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
	//if !ctx.GlobalBool(flags.GetFlagName(utils.RPCLocalEnableFlag)) {
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
