package main

import (
	"fmt"
	"os"
	"runtime"

	chain "github.com/oniio/oniChain/bin"
	chaincmd "github.com/oniio/oniChain/cmd"
	chainutils "github.com/oniio/oniChain/cmd/utils"
	"github.com/oniio/oniChain/common/log"
	"github.com/oniio/oniDNS/cmd"
	"github.com/oniio/oniDNS/common"
	"github.com/oniio/oniDNS/config"
	"github.com/oniio/oniDNS/messageBus"
	"github.com/oniio/oniDNS/netserver"
	"github.com/oniio/oniDNS/storage"
	"github.com/urfave/cli"
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
	chain.StartOntology(ctx)
	//seed
	initLog(ctx)

	err := initConfig(ctx)
	if err != nil {
		log.Errorf("initConfig error:%s", err)
		return
	}
	svr := netserver.NewNetServer()

	if err = svr.Run(); err != nil {
		log.Errorf("run ddns server error:%s", err)
		return
	}
	if err = InitMsgBus(); err != nil {
		log.Fatalf("InitMsgBus error: %v", err)
	}
	storage.TDB, err = InitTrackerDB(TRACKER_DB_PATH)
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

func InitTrackerDB(path string) (*storage.LevelDBStore, error) {
	db, err := storage.NewLevelDBStore(path)
	if err != nil {
		return nil, err
	}
	log.Info("Tracker DB init success")
	return db, nil
}

func InitMsgBus() error {
	messageBus.MsgBus = messageBus.NewMsgBus()
	return messageBus.MsgBus.Start()
}
