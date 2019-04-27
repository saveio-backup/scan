/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-22
 */
package cmd

//import (
//	"encoding/hex"
//	"github.com/oniio/oniDNS/cmd/utils"
//	"github.com/urfave/cli"
//)
//
//var DDNSCommand = cli.Command{
//	Action:      cli.ShowSubcommandHelp,
//	Name:        "ddns",
//	Usage:       "ddns node option:[reg|unReg|req|update]; channel option:[openChannel | listChannel | balance query]",
//	Description: "Local ddns node and channel management",
//	Subcommands: []cli.Command{
//
//		{
//			Action:    registerDns,
//			Name:      "registerdns",
//			Usage:     "Register dns candidate",
//			ArgsUsage: " ",
//			Flags: []cli.Flag{
//				utils.DnsIpFlag,
//				utils.DnsPortFlag,
//				utils.InitDepositFlag,
//			},
//			Description: "Request register as dns candidate",
//		},
//		{
//			Action:      unregisterDns,
//			Name:        "unregisterdns",
//			Usage:       "Cancel previous register request",
//			ArgsUsage:   " ",
//			Description: "Cancel previous register request",
//		},
//		{
//			Action:      quitDns,
//			Name:        "quitdns",
//			Usage:       "Quit working as dns",
//			ArgsUsage:   " ",
//			Description: "Quit working as dns",
//		},
//		{
//			Action:    addPos,
//			Name:      "addInitPos",
//			Usage:     "Increase init deposit",
//			ArgsUsage: " ",
//			Flags: []cli.Flag{
//				utils.DeltaDepositFlag,
//			},
//			Description: "Increase init deposit",
//		},
//		{
//			Action:    reducePos,
//			Name:      "reduceInitPos",
//			Usage:     "Reduce init deposit",
//			ArgsUsage: " ",
//			Flags: []cli.Flag{
//				utils.DeltaDepositFlag,
//			},
//			Description: "Reduce init deposit",
//		},
//		{
//			Action:    getRegisterInfo,
//			Name:      "getRegInfo",
//			Usage:     "Display all or specified Dns register info",
//			ArgsUsage: " ",
//			Flags: []cli.Flag{
//				utils.DnsAllFlag,
//				utils.PeerPubkeyFlag,
//			},
//			Description: "Display all or specified Dns register info",
//		},
//		{
//			Action:    getHostInfo,
//			Name:      "getHostInfo",
//			Usage:     "Display all or specified Dns host info including ip, port",
//			ArgsUsage: " ",
//			Flags: []cli.Flag{
//				utils.DnsAllFlag,
//				utils.DnsWalletFlag,
//			},
//			Description: "Display all or specified Dns host info including ip, port",
//		},
//	},
//}
//
////ddns command
//
//func registerDns(ctx *cli.Context) error {
//	if ctx.NumFlags() < 3 {
//		PrintErrorMsg("Missing argument.")
//		cli.ShowSubcommandHelp(ctx)
//		return nil
//	}
//
//	endpoint, err := dsp.Init(config.Parameters.BaseConfig.WalletDir, config.Parameters.BaseConfig.WalletPwd)
//	if err != nil {
//		PrintErrorMsg("init dsp err:%s\n", err)
//		return err
//	}
//
//	ip := ctx.String(flags.GetFlagName(flags.DnsIpFlag))
//	port := ctx.String(flags.GetFlagName(flags.DnsPortFlag))
//	initDeposit := ctx.Uint64(flags.GetFlagName(flags.InitDepositFlag))
//
//	tx, err := endpoint.Chain.Native.Dns.DNSNodeReg([]byte(ip), []byte(port), initDeposit)
//	if err != nil {
//		PrintErrorMsg("Register candidate err:%s\n", err)
//		return nil
//	}
//	PrintInfoMsg("RegisterCandidate Success")
//	PrintInfoMsg("tx :%s\n", hex.EncodeToString(ccom.ToArrayReverse(tx.ToArray())))
//
//	return nil
//}
//
//func unregisterDns(ctx *cli.Context) error {
//	endpoint, err := dsp.Init(config.Parameters.BaseConfig.WalletDir, config.Parameters.BaseConfig.WalletPwd)
//	if err != nil {
//		PrintErrorMsg("init dsp err:%s\n", err)
//		return err
//	}
//
//	tx, err := endpoint.Chain.Native.Dns.UnregisterDNSNode()
//	if err != nil {
//		PrintErrorMsg("Unregister candidate err:%s\n", err)
//		return nil
//	}
//	PrintInfoMsg("UnregisterCandidate Success")
//	PrintInfoMsg("tx :%s\n", hex.EncodeToString(ccom.ToArrayReverse(tx.ToArray())))
//
//	return nil
//}
//
//func quitDns(ctx *cli.Context) error {
//	endpoint, err := dsp.Init(config.Parameters.BaseConfig.WalletDir, config.Parameters.BaseConfig.WalletPwd)
//	if err != nil {
//		PrintErrorMsg("init dsp err:%s\n", err)
//		return err
//	}
//
//	tx, err := endpoint.Chain.Native.Dns.QuitNode()
//	if err != nil {
//		PrintErrorMsg("Quit candidate err:%s\n", err)
//		return nil
//	}
//	PrintInfoMsg("Quit candidate Success")
//	PrintInfoMsg("tx :%s\n", hex.EncodeToString(ccom.ToArrayReverse(tx.ToArray())))
//
//	return nil
//}
//
//func addPos(ctx *cli.Context) error {
//	if ctx.NumFlags() < 1 {
//		PrintErrorMsg("Missing argument.")
//		cli.ShowSubcommandHelp(ctx)
//		return nil
//	}
//
//	endpoint, err := dsp.Init(config.Parameters.BaseConfig.WalletDir, config.Parameters.BaseConfig.WalletPwd)
//	if err != nil {
//		PrintErrorMsg("init dsp err:%s\n", err)
//		return err
//	}
//
//	deltaDeposit := ctx.Uint64(flags.GetFlagName(flags.DeltaDepositFlag))
//
//	tx, err := endpoint.Chain.Native.Dns.AddInitPos(deltaDeposit)
//	if err != nil {
//		PrintErrorMsg("Add init deposit err:%s\n", err)
//		return nil
//	}
//	PrintInfoMsg("Add init deposit Success")
//	PrintInfoMsg("tx :%s\n", hex.EncodeToString(ccom.ToArrayReverse(tx.ToArray())))
//
//	return nil
//}
//
//func reducePos(ctx *cli.Context) error {
//	if ctx.NumFlags() < 1 {
//		PrintErrorMsg("Missing argument.")
//		cli.ShowSubcommandHelp(ctx)
//		return nil
//	}
//
//	endpoint, err := dsp.Init(config.Parameters.BaseConfig.WalletDir, config.Parameters.BaseConfig.WalletPwd)
//	if err != nil {
//		PrintErrorMsg("init dsp err:%s\n", err)
//		return err
//	}
//
//	deltaDeposit := ctx.Uint64(flags.GetFlagName(flags.DeltaDepositFlag))
//
//	tx, err := endpoint.Chain.Native.Dns.ReduceInitPos(deltaDeposit)
//	if err != nil {
//		PrintErrorMsg("Reduce init deposit err:%s\n", err)
//		return nil
//	}
//	PrintInfoMsg("Reduce init deposit Success")
//	PrintInfoMsg("tx :%s\n", hex.EncodeToString(ccom.ToArrayReverse(tx.ToArray())))
//
//	return nil
//}
//
//func getRegisterInfo(ctx *cli.Context) error {
//	endpoint, err := dsp.Init(config.Parameters.BaseConfig.WalletDir, config.Parameters.BaseConfig.WalletPwd)
//	if err != nil {
//		PrintErrorMsg("init dsp err:%s\n", err)
//		return err
//	}
//
//	DnsAllFlag := ctx.Bool(flags.GetFlagName(flags.DnsAllFlag))
//
//	if DnsAllFlag {
//		m, err := endpoint.Chain.Native.Dns.GetPeerPoolMap()
//
//		if err != nil {
//			PrintErrorMsg("Get all dns register info err:%s\n", err)
//			return nil
//		}
//
//		if _, ok := m.PeerPoolMap[""]; ok {
//			delete(m.PeerPoolMap, "")
//		}
//
//		for _, item := range m.PeerPoolMap {
//			PrintInfoMsg("PeerPubkey: %s\n", item.PeerPubkey)
//			PrintInfoMsg("WalletAddress: %s\n", item.WalletAddress.ToBase58())
//			PrintInfoMsg("Status: %d\n", item.Status)
//			PrintInfoMsg("InitPos: %d\n", item.TotalInitPos)
//			PrintInfoMsg("\n")
//		}
//	} else {
//		peerPubkey := ctx.String(flags.GetFlagName(flags.PeerPubkeyFlag))
//		item, err := endpoint.Chain.Native.Dns.GetPeerPoolItem(peerPubkey)
//		if err != nil {
//			PrintErrorMsg("Get dns register info err:%s\n", err)
//			return nil
//		}
//
//		PrintInfoMsg("PeerPubkey: %s\n", item.PeerPubkey)
//		PrintInfoMsg("WalletAddress: %s\n", item.WalletAddress.ToBase58())
//		PrintInfoMsg("Status: %d\n", item.Status)
//		PrintInfoMsg("TotalInitPos: %d\n", item.TotalInitPos)
//	}
//
//	return nil
//}
//
//func getHostInfo(ctx *cli.Context) error {
//	endpoint, err := dsp.Init(config.Parameters.BaseConfig.WalletDir, config.Parameters.BaseConfig.WalletPwd)
//	if err != nil {
//		PrintErrorMsg("init dsp err:%s\n", err)
//		return err
//	}
//
//	DnsAllFlag := ctx.Bool(flags.GetFlagName(flags.DnsAllFlag))
//
//	if DnsAllFlag {
//		infos, err := endpoint.Chain.Native.Dns.GetAllDnsNodes()
//		if err != nil {
//			PrintErrorMsg("Get all dns host info err:%s\n", err)
//			return nil
//		}
//
//		for k, v := range infos {
//			PrintInfoMsg("Pubkey:%s\n", k)
//			PrintInfoMsg("wallet:%s\n", v.WalletAddr.ToBase58())
//			PrintInfoMsg("ip:%s\n", v.IP)
//			PrintInfoMsg("port:%s\n", v.Port)
//			PrintInfoMsg("\n")
//		}
//	} else {
//		var addr ccom.Address
//		var err error
//
//		walletAddr := ctx.String(flags.GetFlagName(flags.DnsWalletFlag))
//		if walletAddr != "" {
//			addr, err = ccom.AddressFromBase58(walletAddr)
//			if err != nil {
//				PrintErrorMsg("Get dns host info err:%s\n", err)
//				return nil
//			}
//		}
//
//		info, err := endpoint.Chain.Native.Dns.GetDnsNodeByAddr(addr)
//		if err != nil {
//			PrintErrorMsg("Get dns host info err:%s\n", err)
//			return nil
//		}
//
//		PrintInfoMsg("Pubkey:%s\n", info.PeerPubKey)
//		PrintInfoMsg("Wallet:%s\n", info.WalletAddr.ToBase58())
//		PrintInfoMsg("Ip:%v\n", info.IP)
//		PrintInfoMsg("Port:%v\n", info.Port)
//	}
//
//	return nil
//}
