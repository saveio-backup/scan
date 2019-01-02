package cmd

import (
	"github.com/oniio/oniDNS/config"
	"github.com/urfave/cli"
)

func SetSeedsConfig(ctx *cli.Context) (*config.CommonConfig, error) {
	cfg := config.DefaultConfig
	setCommonConfig(ctx, cfg)
	return cfg, nil
}

func setCommonConfig(ctx *cli.Context, cfg *config.CommonConfig) {
	cfg.LogLevel = ctx.Uint(GetFlagName(LogLevelFlag))
	cfg.LogStderr = ctx.Bool(GetFlagName(LogStderrFlag))

}
