package utils

import (
	"strings"

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
	}

	RpcServerFlag = cli.StringFlag{
		Name:  "rpcServer",
		Usage: "",
		Value: "",
	}
	ConfPathFlag = cli.StringFlag{
		Name:  "confPath",
		Usage: "",
		Value: "",
	}

	DbDirFlag = cli.StringFlag{
		Name:  "db-dir",
		Usage: "db `<path>`",
	}

	//p2p command setting
	NetworkIDFlag = cli.UintFlag{
		Name:  "networkId",
		Usage: "",
	}

	PortFlag = cli.UintFlag{
		Name:  "p2pPort",
		Usage: "P2P network port `<number>`",
	}
	SeedListFlag = cli.StringSliceFlag{
		Name:  "seedlist",
		Usage: "P2P network seedlist `<protocol://ip:port>`",
	}

	//tracker command setting
	TrackerServerPortFlag = cli.UintFlag{
		Name:  "trackerport",
		Usage: "tracker server listen udp port `<number>`",
	}
	TrackerFee = cli.Uint64Flag{
		Name:  "trackerfee",
		Usage: "tracker fee `<uint>`",
	}
	WalletFlag = cli.StringFlag{
		Name:  "walletstring,ws",
		Usage: "wallet address of base58 format",
	}
	WalletSetFlag = cli.BoolFlag{
		Name:   "unsetWallet",
		Usage:  "Enable restful api server",
		Hidden: false,
	}
	WalletFileFlag = cli.StringFlag{
		Name:  "wallet,w",
		Usage: "Wallet `<file>`",
	}
	HostFlag = cli.StringFlag{
		Name:  "host",
		Usage: "ip address and port of string format",
	}

	//ddns command setting
	DnsIpFlag = cli.StringFlag{
		Name:  "dnsIp",
		Usage: "Dns `<ip>`",
	}
	DnsPortFlag = cli.StringFlag{
		Name:  "dnsPort",
		Usage: "Dns `<port>`",
	}
	DnsWalletFlag = cli.StringFlag{
		Name:  "walletAddr",
		Usage: "Dns `<walletAddr>`",
	}
	DnsAllFlag = cli.BoolFlag{
		Name:  "all",
		Usage: "All Dns info",
	}
	//channel command setting
	PartnerAddressFlag = cli.StringFlag{
		Name:  "partnerAddr",
		Usage: "Channel partner `<address>`",
	}
	TargetAddressFlag = cli.StringFlag{
		Name:  "targetAddr",
		Usage: "Channel transfer target `<address>`",
	}
	TotalDepositFlag = cli.Uint64Flag{
		Name:  "totalDeposit",
		Usage: "Channel total `<deposit>`",
	}
	AmountFlag = cli.Uint64Flag{
		Name:  "amount",
		Usage: "Channel payment amount `<amount>`",
	}
	PaymentIDFlag = cli.UintFlag{
		Name:  "paymentId",
		Usage: "",
	}

	//ddns govern command setting
	PeerPubkeyFlag = cli.StringFlag{
		Name:  "peerPubkey",
		Usage: "candidate pubkey",
	}
	InitDepositFlag = cli.Uint64Flag{
		Name:  "initDeposit",
		Usage: "Init `<deposit>`",
	}
	PeerPubkeyListFlag = cli.StringFlag{
		Name:  "peerPubkeyList",
		Usage: "candidate pubkey list",
	}
	WithdrawListFlag = cli.StringFlag{
		Name:  "withdrawList",
		Usage: "withdraw value list",
	}
	DeltaDepositFlag = cli.Uint64Flag{
		Name:  "deltaDeposit",
		Usage: "Delta `<deposit>`",
	}
	// RPC settings
	RPCDisabledFlag = cli.BoolFlag{
		Name:  "disable-rpc",
		Usage: "Shut down the rpc server.",
	}
	RPCPortFlag = cli.UintFlag{
		Name:  "rpcport",
		Usage: "Json rpc server listening port `<number>`",
	}
	RPCLocalEnableFlag = cli.BoolFlag{
		Name:  "localrpc",
		Usage: "Enable local rpc server",
	}
	RPCLocalProtFlag = cli.UintFlag{
		Name:  "localrpcport",
		Usage: "Json rpc local server listening port `<number>`",
	}

	//Restful setting
	RestfulEnableFlag = cli.BoolFlag{
		Name:  "rest",
		Usage: "Enable restful api server",
	}
	RestfulPortFlag = cli.UintFlag{
		Name:  "restport",
		Usage: "Restful server listening port `<number>`",
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
