/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-26
 */
package cmd

import (
	"github.com/saveio/themis/common/log"
	"github.com/saveio/scan/cmd/utils"
	"github.com/saveio/scan/common/config"
	alog "github.com/ontio/ontology-eventbus/log"
	"github.com/urfave/cli"
	"os"
)

//Modify default common config by cli
func setCommonConfig(cfg *config.CommonConfig, ctx *cli.Context) {
	if ctx.GlobalIsSet(utils.GetFlagName(utils.DbDirFlag)) {
		cfg.CommonDBPath = ctx.String(utils.GetFlagName(utils.DbDirFlag))
	}
	if ctx.GlobalIsSet(utils.GetFlagName(utils.ConfPathFlag)) {
		cfg.ConfPath = ctx.String(utils.GetFlagName(utils.ConfPathFlag))
	}

}

//Modify default p2p config by cli
func setP2PnodeConfig(cfg *config.P2PConfig, ctx *cli.Context) {
	if ctx.GlobalIsSet(utils.GetFlagName(utils.NetworkIDFlag)) {
		cfg.NetworkId = uint32(ctx.Uint(utils.GetFlagName(utils.NetworkIDFlag)))
	}
	if ctx.GlobalIsSet(utils.GetFlagName(utils.PortFlag)) {
		cfg.PortBase = ctx.Uint(utils.GetFlagName(utils.PortFlag))
	}
	if ctx.GlobalIsSet(utils.GetFlagName(utils.SeedListFlag)) {
		cfg.SeedList = ctx.StringSlice(utils.GetFlagName(utils.SeedListFlag))
	}

}

func setTrackerConfig(cfg *config.TrackerConfig, ctx *cli.Context) {
	if ctx.GlobalIsSet(utils.GetFlagName(utils.TrackerServerPortFlag)) {
		cfg.UdpPort = ctx.Uint(utils.GetFlagName(utils.TrackerServerPortFlag))
	}
	if ctx.GlobalIsSet(utils.GetFlagName(utils.TrackerFee)) {
		cfg.Fee = ctx.Uint64(utils.GetFlagName(utils.TrackerFee))
	}
}

///////////////////////set channel service config by cli///////////////////////

func SetLogConfig(ctx *cli.Context) {
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

}

func setRpcConfig(cfg *config.RpcConfig, ctx *cli.Context) {
	if ctx.GlobalIsSet(utils.GetFlagName(utils.RPCDisabledFlag)) {
		cfg.EnableHttpJsonRpc = !ctx.Bool(utils.GetFlagName(utils.RPCDisabledFlag))
	}
	if ctx.GlobalIsSet(utils.GetFlagName(utils.RPCPortFlag)) {
		cfg.HttpJsonPort = ctx.Uint(utils.GetFlagName(utils.RPCPortFlag))
	}

	if ctx.GlobalIsSet(utils.GetFlagName(utils.RPCLocalProtFlag)) {
		cfg.HttpLocalPort = ctx.Uint(utils.GetFlagName(utils.RPCLocalProtFlag))
	}
}

func setRestfulConfig(cfg *config.RestfulConfig, ctx *cli.Context) {
	if ctx.GlobalIsSet(utils.GetFlagName(utils.RestfulEnableFlag)) {
		cfg.EnableHttpRestful = ctx.Bool(utils.GetFlagName(utils.RestfulEnableFlag))
	}
	if ctx.GlobalIsSet(utils.GetFlagName(utils.RestfulPortFlag)) {
		cfg.HttpRestPort = ctx.Uint(utils.GetFlagName(utils.RestfulPortFlag))
	}

}

func SetRpcPort(ctx *cli.Context) {
	if ctx.GlobalIsSet(utils.GetFlagName(utils.RPCPortFlag)) {
		config.DefaultConfig.RpcConfig.HttpJsonPort = ctx.Uint(utils.GetFlagName(utils.RPCPortFlag))
	}
}

func SetDDNSConfig(ctx *cli.Context) {
	DefaultConfig := config.DefaultConfig
	setCommonConfig(&DefaultConfig.CommonConfig, ctx)
	setP2PnodeConfig(&DefaultConfig.P2PConfig, ctx)
	setTrackerConfig(&DefaultConfig.TrackerConfig, ctx)
	setRpcConfig(&DefaultConfig.RpcConfig, ctx)
	setRestfulConfig(&DefaultConfig.RestfulConfig, ctx)
}

func Init(ctx *cli.Context) {
	if ctx.GlobalIsSet(utils.GetFlagName(utils.ConfPathFlag)) {
		config.ConfigDir = ctx.String(utils.GetFlagName(utils.ConfPathFlag)) + config.DEFAULT_CONFIG_FILE
	} else {
		config.ConfigDir = os.Getenv("HOME") + config.DEFAULT_CONFIG_FILE
	}
	config.GetJsonObjectFromFile(config.ConfigDir, config.DefaultConfig)
}
