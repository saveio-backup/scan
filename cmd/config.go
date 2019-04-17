package cmd

import (
	"github.com/oniio/oniDNS/config"
	"github.com/urfave/cli"
	"github.com/oniio/oniDNS/cmd/utils"
)

func SetSeedsConfig(ctx *cli.Context) (*config.CommonConfig, error) {
	cfg := config.DefaultConfig
	setCommonConfig(ctx, cfg)
	return cfg, nil
}

func setCommonConfig(ctx *cli.Context, cfg *config.CommonConfig) {
	cfg.LogLevel = ctx.Uint(utils.GetFlagName(utils.LogLevelFlag))
	cfg.LogStderr = ctx.Bool(utils.GetFlagName(utils.LogStderrFlag))

}

func setRpcConfig(ctx *cli.Context, cfg *config.RpcConfig) {
	cfg.EnableHttpJsonRpc = !ctx.Bool(utils.GetFlagName(utils.RPCDisabledFlag))
	cfg.HttpJsonPort = ctx.Uint(utils.GetFlagName(utils.RPCPortFlag))
	cfg.HttpLocalPort = ctx.Uint(utils.GetFlagName(utils.RPCLocalProtFlag))
}

func setRestfulConfig(ctx *cli.Context, cfg *config.RestfulConfig) {
	cfg.EnableHttpRestful = ctx.Bool(utils.GetFlagName(utils.RestfulEnableFlag))
	cfg.HttpRestPort = ctx.Uint(utils.GetFlagName(utils.RestfulPortFlag))
}

func SetRpcPort(ctx *cli.Context) {
	if ctx.IsSet(utils.GetFlagName(utils.RPCPortFlag)) {
		config.DefaultConfig.Rpc.HttpJsonPort = ctx.Uint(utils.GetFlagName(utils.RPCPortFlag))
	}
}


