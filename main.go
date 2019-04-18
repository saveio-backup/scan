package main

import (
	"fmt"
	"os"
	"runtime"

	chaincmd "github.com/oniio/oniChain/cmd"
	chainutils "github.com/oniio/oniChain/cmd/utils"
	"github.com/oniio/oniChain/common/log"

	chain "github.com/oniio/oniChain/start"
	"github.com/oniio/oniDNS/cmd"
	"github.com/oniio/oniDNS/common"
	"github.com/oniio/oniDNS/config"
	"github.com/oniio/oniDNS/netserver"
	"github.com/oniio/oniDNS/network"
	"github.com/oniio/oniDNS/network/actor/recv"
	"github.com/oniio/oniDNS/storage"
	tcomm "github.com/oniio/oniDNS/tracker/common"
	"github.com/ontio/ontology-eventbus/actor"
	alog "github.com/ontio/ontology-eventbus/log"
	"github.com/urfave/cli"
	//"github.com/oniio/oniChain-go-sdk/wallet"
	"github.com/oniio/oniDNS/tracker"
	"github.com/oniio/oniDNS/cmd/utils"
	"github.com/oniio/oniDNS/http/localrpc"
	"time"
	"github.com/oniio/oniDNS/http/restful"
	"github.com/oniio/oniDNS/http/jsonrpc"
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
	}
	app.Flags = []cli.Flag{
		//common setting
		utils.LogStderrFlag,
		utils.LogLevelFlag,
		utils.HostFlag,
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
		chainutils.WsEnabledFlag,
		chainutils.WsPortFlag,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	return app
}

func main() {
	if err := initAPP().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func startDDNS(ctx *cli.Context) {
	//chainnode
	err := startChain(ctx)
	if err != nil {
		log.Error("Chain init error")
		return
	}
	startTracker(ctx)
	common.WaitToExit()
}

func startTracker(ctx *cli.Context){
	initLog(ctx)
	err := initConfig(ctx)
	if err != nil {
		log.Errorf("initConfig error:%s", err)
		return
		}
	p2pSvr, p2pPid, err := initP2P()
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
	recv.P2pPid = p2pPid
	network.DDNSP2P = p2pSvr
	storage.TDB, err = initTrackerDB(config.TRACKER_DB_PATH)
	if err != nil {
		log.Fatalf("InitTrackerDB error: %v", err)
	}

	svr.Tsvr.SetPID(p2pPid)
	if err = svr.Run(); err != nil {
		log.Errorf("run ddns server error:%s", err)
		return
	}
	err = initRpc(ctx)
	if err != nil {
		log.Errorf("initRpc error:%s", err)
		return
	}
	initRestful(ctx)
	//host:=p2pSvr.ExternalAddr
	//w, err := wallet.OpenWallet(config.WALLET_FILE)
	//acc,err:=w.GetDefaultAccount([]byte(config.WALLET_PWD))
	//wAddr:=acc.Address
	//ws:=wAddr.ToHexString()
	ws:=config.Wallet1Addr
	if err != nil {
		log.Errorf("open wallet err:%s\n", err)
		return
	}
	//TODO: to check if had registerd in contract
	//temp test
	if err=tracker.EndPointRegistry(ws,config.Host);err!=nil{
		log.Errorf("DDNS node EndPointRegistry error:%v",err)
	}
	log.Info("Tracker start success")
}

func initConfig(ctx *cli.Context) error {
	_, err := cmd.SetSeedsConfig(ctx)
	if err != nil {
		return err
	}
	log.Info("Config init success")
	return nil
}

func initLog(ctx *cli.Context) {
	//init log module
	if ctx.Bool(utils.GetFlagName(utils.LogStderrFlag)) {
		logLevel := ctx.GlobalInt(utils.GetFlagName(utils.LogLevelFlag))
		log.InitLog(logLevel, log.Stdout)
	} else {
		logLevel := ctx.GlobalInt(utils.GetFlagName(utils.LogLevelFlag))
		alog.InitLog(log.PATH)
		log.InitLog(logLevel, log.PATH, log.Stdout)
	}
	//log.SetLevel(ctx.GlobalUint(cmd.GetFlagName(cmd.LogLevelFlag))) //TODO
	//log.SetMaxSize(config.DEFAULT_MAX_LOG_SIZE) //TODO
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

func initP2P() (*network.Network, *actor.PID, error) {
	log.Info("P2P is init...")
	p2p := network.NewP2P()
	go p2p.Start()
	pid, err := recv.NewP2PActor(p2p)
	BlockUntilComplete()
	if err != nil {
		return nil, nil, fmt.Errorf("p2pActor init error %s", err)
	}
	p2p.SetPID(pid)
	log.Infof("PID:%v", pid)
	log.Info("p2p init success")
	return p2p, pid, nil
}

func initRpc(ctx *cli.Context) error {
	if !config.DefaultConfig.Rpc.EnableHttpJsonRpc {
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
	log.Infof("Rpc init success")
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
	if !config.DefaultConfig.Restful.EnableHttpRestful {
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
