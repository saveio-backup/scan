package cmd

import (
	"os"

	alog "github.com/ontio/ontology-eventbus/log"
	"github.com/saveio/scan/cmd/utils"
	"github.com/saveio/scan/common/config"
	"github.com/saveio/themis/common/log"
	"github.com/urfave/cli"
)

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

func SetRpcPort(ctx *cli.Context) {
	if ctx.GlobalIsSet(utils.GetFlagName(utils.RPCPortFlag)) {
		config.Parameters.Base.JsonRpcPortOffset = ctx.Uint(utils.GetFlagName(utils.RPCPortFlag))
	}
}

func SetDDNSConfig(ctx *cli.Context) {
}

func Init(ctx *cli.Context) {
	if ctx.GlobalIsSet(utils.GetFlagName(utils.ConfPathFlag)) {
		config.ConfigDir = ctx.String(utils.GetFlagName(utils.ConfPathFlag)) + config.DEFAULT_CONFIG_FILE
	} else {
		config.ConfigDir = os.Getenv("HOME") + config.DEFAULT_CONFIG_FILE
	}
	config.GetJsonObjectFromFile(config.ConfigDir, config.Parameters)
}
