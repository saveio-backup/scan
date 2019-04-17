package utils

import (
	"strings"

	"github.com/oniio/oniDNS/config"
	"github.com/urfave/cli"
)

var (
	//commmon
	LogStderrFlag = cli.BoolFlag{
		Name:  "logstderr",
		Usage: "log to standard error instead of files,default false",
	}
	LogLevelFlag = cli.UintFlag{
		Name:  "loglevel",
		Usage: "Set the log level to `<level>` (0~4). 0:DEBUG 1:INFO 2:WARNING 3:ERROR 4:FATAL",
		Value: config.DEFAULT_LOG_LEVEL,
	}
	WalletFlag = cli.StringFlag{
		Name:"wallet,w",
		Usage:"wallet address of string format",
	}
	WalletFileFlag = cli.StringFlag{
		Name:  "wallet,w",
		Value: config.DEFAULT_WALLET_FILE_NAME,
		Usage: "Wallet `<file>`",
	}
	HostFlag = cli.StringFlag{
		Name:"host",
		Usage:"ip address and port of string format",
		Value:config.DEFAULT_HOST,
	}
	// RPC settings
	RPCDisabledFlag = cli.BoolFlag{
		Name:  "disable-rpc",
		Usage: "Shut down the rpc server.",
	}
	RPCPortFlag = cli.UintFlag{
		Name:  "rpcport",
		Usage: "Json rpc server listening port `<number>`",
		Value: config.DEFAULT_RPC_PORT,
	}
	RPCLocalEnableFlag = cli.BoolFlag{
		Name:  "localrpc",
		Usage: "Enable local rpc server",
	}
	RPCLocalProtFlag = cli.UintFlag{
		Name:  "localrpcport",
		Usage: "Json rpc local server listening port `<number>`",
		Value: config.DEFAULT_RPC_LOCAL_PORT,
	}
	
	//Restful setting
	RestfulEnableFlag = cli.BoolFlag{
		Name:  "rest",
		Usage: "Enable restful api server",
	}
	RestfulPortFlag = cli.UintFlag{
		Name:  "restport",
		Usage: "Restful server listening port `<number>`",
		Value: config.DEFAULT_REST_PORT,
	}
	
)

//GetFlagName deal with short flag, and return the flag name whether flag name have short name
func GetFlagName(flag cli.Flag) string {
	name := flag.GetName()
	if name == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(name, ",")[0])
}
