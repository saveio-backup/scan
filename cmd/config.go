package cmd

import (
	"os"

	alog "github.com/ontio/ontology-eventbus/log"
	"github.com/saveio/scan/cmd/flags"
	"github.com/saveio/scan/common/config"
	"github.com/saveio/themis/common/log"
	"github.com/urfave/cli"
)

///////////////////////set channel service config by cli///////////////////////

func InitLog(ctx *cli.Context) {
	if ctx.Bool(flags.GetFlagName(flags.LogStderrFlag)) {
		logLevel := ctx.GlobalInt(flags.GetFlagName(flags.LogLevelFlag))
		log.InitLog(logLevel, log.Stdout)
	} else {
		logLevel := ctx.GlobalInt(flags.GetFlagName(flags.LogLevelFlag))
		alog.InitLog(log.PATH)
		log.InitLog(logLevel, log.PATH, log.Stdout)
		log.AddIgnore("pylons")
		log.AddIgnore("transfer")
	}
	//log.SetLevel(ctx.GlobalUint(cmd.GetFlagName(cmd.LogLevelFlag))) //TODO
	//log.SetMaxSize(config.DEFAULT_MAX_LOG_SIZE) //TODO
}

func SetRpcPort(ctx *cli.Context) {
	if ctx.GlobalIsSet(flags.GetFlagName(flags.RPCPortFlag)) {
		config.Parameters.Base.JsonRpcPortOffset = ctx.Uint(flags.GetFlagName(flags.RPCPortFlag))
	}
}

func SetDDNSConfig(ctx *cli.Context) {
}

func Init(ctx *cli.Context) {
	if ctx.GlobalIsSet(flags.GetFlagName(flags.ConfPathFlag)) {
		config.ConfigDir = ctx.String(flags.GetFlagName(flags.ConfPathFlag)) + config.DEFAULT_CONFIG_FILE
	} else {
		config.ConfigDir = os.Getenv("HOME") + config.DEFAULT_CONFIG_FILE
	}

	config.GetJsonObjectFromFile(config.DEFAULT_CONFIG_FILE, config.Parameters)
}
