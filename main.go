package main

import (
	"fmt"
	"os"
	"runtime"

	chaincmd "github.com/oniio/oniChain/cmd"
	chainutils "github.com/oniio/oniChain/cmd/utils"
	"github.com/oniio/oniChain/common/log"
	//chain "github.com/oniio/oniChain/start"
	"github.com/oniio/oniDNS/cmd"
	"github.com/oniio/oniDNS/common"
	"github.com/oniio/oniDNS/config"
	"github.com/oniio/oniDNS/netserver"
	"github.com/oniio/oniDNS/storage"
	"github.com/urfave/cli"
	"github.com/oniio/oniDNS/network"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/oniio/oniDNS/network/actor/recv"
	tcomm "github.com/oniio/oniDNS/tracker/common"
)

var (
	TRACKER_DB_PATH = "./TrackerLevelDB"
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
	}
	app.Flags = []cli.Flag{
		//common setting
		cmd.LogStderrFlag,
		cmd.LogLevelFlag,
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
	//chain.StartOntology(ctx)
	//seed
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
	recv.P2pPid=p2pPid
	network.DDNSP2P=p2pSvr
	storage.TDB, err = initTrackerDB(TRACKER_DB_PATH)
	if err != nil {
		log.Fatalf("InitTrackerDB error: %v", err)
	}

	common.WaitToExit()
}

func initConfig(ctx *cli.Context) error {
	_, err := cmd.SetSeedsConfig(ctx)
	if err != nil {
		return err
	}
	log.Infof("Config init success")
	return nil
}

func initLog(ctx *cli.Context) {
	//init log module
	//log.SetLevel(ctx.GlobalUint(cmd.GetFlagName(cmd.LogLevelFlag))) //TODO
	//log.SetMaxSize(config.DEFAULT_MAX_LOG_SIZE) //TODO
	if ctx.Bool(cmd.GetFlagName(cmd.LogStderrFlag)) {
		log.InitLog(0, config.DEFAULT_LOG_DIR)
	} else {
		log.InitLog(1, config.DEFAULT_LOG_DIR)
	}
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

func initP2P()(*network.Network,*actor.PID,error){
	log.Info("P2P is init...")
	p2p:=network.NewP2P()
	go p2p.Start()
	pid,err:=recv.NewP2PActor(p2p)
	if err != nil {
		return nil, nil, fmt.Errorf("p2pActor init error %s", err)
	}
	p2p.SetPID(pid)
	log.Infof("PID:%v",pid)
	log.Info("p2p init success")
	return p2p,pid,nil
}

func BlockUntilComplete() {
	<-tcomm.ListeningCh
}


