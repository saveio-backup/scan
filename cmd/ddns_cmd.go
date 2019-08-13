package cmd

import (
	"github.com/saveio/scan/cmd/common"
	"github.com/saveio/scan/cmd/flags"
	"github.com/saveio/scan/cmd/utils"
	httpComm "github.com/saveio/scan/http/base/common"
	"github.com/urfave/cli"
)

var DNSCommand = cli.Command{
	Action:      cli.ShowSubcommandHelp,
	Name:        "dns",
	Usage:       "Interactive with native dns contract",
	Description: "Interactive with native dns contract",
	Subcommands: []cli.Command{

		{
			Action:    registerDns,
			Name:      "reg",
			Usage:     "Register dns candidate",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.DnsIpFlag,
				flags.DnsPortFlag,
				flags.InitDepositFlag,
			},
			Description: "Request register as dns candidate",
		},
		{
			Action:      unregisterDns,
			Name:        "unreg",
			Usage:       "Cancel previous register request",
			ArgsUsage:   " ",
			Description: "Cancel previous register request",
		},
		{
			Action:      quitDns,
			Name:        "quit",
			Usage:       "Quit working as dns",
			ArgsUsage:   " ",
			Description: "Quit working as dns",
		},
		{
			Action:    addPos,
			Name:      "addPos",
			Usage:     "Increase init deposit",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.DeltaDepositFlag,
			},
			Description: "Increase init deposit",
		},
		{
			Action:    reducePos,
			Name:      "reducePos",
			Usage:     "Reduce init deposit",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.DeltaDepositFlag,
			},
			Description: "Reduce init deposit",
		},
		{
			Action:    getRegisterInfo,
			Name:      "getRegInfo",
			Usage:     "Display all or specified Dns register info",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.DnsAllFlag,
				flags.PeerPubkeyFlag,
			},
			Description: "Display all or specified Dns register info",
		},
		{
			Action:    getHostInfo,
			Name:      "getHostInfo",
			Usage:     "Display all or specified Dns host info including ip, port",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.DnsAllFlag,
				flags.DnsWalletFlag,
			},
			Description: "Display all or specified Dns host info including ip, port",
		},
		{
			Action:      checkTopology,
			Name:        "topology",
			Usage:       "Display dns node local network topology",
			ArgsUsage:   " ",
			Flags:       []cli.Flag{},
			Description: "Display dns node local network topology",
		},
	},
}

//ddns command

func registerDns(ctx *cli.Context) error {
	if ctx.NumFlags() < 3 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("GetPasswd error, ErrMsg:")
		return err
	}
	_, err = client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("GetDefaultAccount from wallet and password error, ErrMsg:")
		return err
	}

	ip := ctx.String(flags.GetFlagName(flags.DnsIpFlag))
	port := ctx.String(flags.GetFlagName(flags.DnsPortFlag))
	initDeposit := ctx.Uint64(flags.GetFlagName(flags.InitDepositFlag))

	dnsRsp, failed := utils.DNSNodeReg(ip, port, initDeposit)

	if failed != nil {
		PrintErrorMsg("\nRegister dns failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}
	PrintInfoMsg("\nRegister dns success. \n\n  Tips: Using \"./ddns info status %s\" to query transaction status.\n", dnsRsp.Tx)
	return nil
}

func unregisterDns(ctx *cli.Context) error {
	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("GetPasswd error, ErrMsg:")
		return err
	}
	_, err = client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("GetDefaultAccount from wallet and password error, ErrMsg:")
		return err
	}

	dnsRsp, failed := utils.DNSNodeUnreg()
	if failed != nil {
		PrintErrorMsg("\nUnregister dns failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}
	PrintInfoMsg("\nUnregister dns success. \n\n  Tips: Using \"./ddns info status %s\" to query transaction status.\n", dnsRsp.Tx)
	return nil
}

func quitDns(ctx *cli.Context) error {
	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("GetPasswd error, ErrMsg:")
		return err
	}
	_, err = client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("GetDefaultAccount from wallet and password error, ErrMsg:")
		return err
	}

	dnsRsp, failed := utils.DNSNodeQuit()
	if failed != nil {
		PrintErrorMsg("\nQuit dns failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}
	PrintInfoMsg("\nQuit dns success. \n\n  Tips: Using \"./ddns info status %s\" to query transaction status.\n", dnsRsp.Tx)
	return nil
}

func addPos(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("GetPasswd error, ErrMsg:")
		return err
	}
	_, err = client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("GetDefaultAccount from wallet and password error, ErrMsg:")
		return err
	}

	deltaDeposit := ctx.Uint64(flags.GetFlagName(flags.DeltaDepositFlag))
	dnsRsp, failed := utils.DNSAddPos(deltaDeposit)
	if failed != nil {
		PrintErrorMsg("\nAdd dns init deposit failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}
	PrintInfoMsg("\nAdd dns init deposit success. \n\n  Tips: Using \"./ddns info status %s\" to query transaction status.\n", dnsRsp.Tx)
	return nil
}

func reducePos(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("GetPasswd error, ErrMsg:")
		return err
	}
	_, err = client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("GetDefaultAccount from wallet and password error, ErrMsg:")
		return err
	}

	deltaDeposit := ctx.Uint64(flags.GetFlagName(flags.DeltaDepositFlag))
	dnsRsp, failed := utils.DNSReducePos(deltaDeposit)
	if failed != nil {
		PrintErrorMsg("\nReduce dns init deposit failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}
	PrintInfoMsg("\nReduce dns init deposit success. \n\n  Tips: Using \"./ddns info status %s\" to query transaction status.\n", dnsRsp.Tx)
	return nil
}

func getRegisterInfo(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("GetPasswd error, ErrMsg:")
		return err
	}
	_, err = client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("GetDefaultAccount from wallet and password error, ErrMsg:")
		return err
	}

	dnsAll := ctx.Bool(flags.GetFlagName(flags.DnsAllFlag))
	peerPubkey := ctx.String(flags.GetFlagName(flags.PeerPubkeyFlag))

	dnsPPRsp, failed := utils.DNSRegisterInfo(dnsAll, peerPubkey)
	if failed != nil {
		PrintErrorMsg("\nGet dns register info failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	} else if dnsPPRsp != nil {
		if dnsPPRsp.PeerPoolMap != nil {
			PrintInfoMsg("\nGet dns register info success. Default limit 50")
			counter := 0
			for _, item := range dnsPPRsp.PeerPoolMap {
				if counter < 50 {
					PrintInfoMsg("Index %d:", counter)
					PrintJsonObject(httpComm.DnsRspPeerPoolItem{
						PeerPubkey:    item.PeerPubkey,
						WalletAddress: item.WalletAddress.ToBase58(),
						Status:        item.Status,
						TotalInitPos:  item.TotalInitPos,
					})
				}
				counter++
			}
			if counter == 0 {
				PrintInfoMsg("Empty")
			}
			return nil
		} else if dnsPPRsp.PeerPoolItem != nil {
			PrintInfoMsg("\nGet dns register info success.")
			PrintJsonObject(httpComm.DnsRspPeerPoolItem{
				PeerPubkey:    dnsPPRsp.PeerPoolItem.PeerPubkey,
				WalletAddress: dnsPPRsp.PeerPoolItem.WalletAddress.ToBase58(),
				Status:        dnsPPRsp.PeerPoolItem.Status,
				TotalInitPos:  dnsPPRsp.PeerPoolItem.TotalInitPos,
			})
			return nil
		}
	} else {
		PrintInfoMsg("\nGet dns register info success. result is null")
		return nil
	}
	return nil
}

func getHostInfo(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("GetPasswd error, ErrMsg:")
		return err
	}
	_, err = client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("GetDefaultAccount from wallet and password error, ErrMsg:")
		return err
	}

	dnsAll := ctx.Bool(flags.GetFlagName(flags.DnsAllFlag))
	walletAddr := ctx.String(flags.GetFlagName(flags.DnsWalletFlag))

	dnsNIRsp, failed := utils.DNSHostInfo(dnsAll, walletAddr)
	if failed != nil {
		PrintErrorMsg("\nGet dns host info failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	} else if dnsNIRsp != nil {
		if dnsNIRsp.NodeInfoMap != nil {
			PrintInfoMsg("\nGet dns host info success. Default limit 50")
			counter := 0
			for _, item := range dnsNIRsp.NodeInfoMap {
				if counter < 50 {
					PrintInfoMsg("Index %d:", counter)
					PrintJsonObject(httpComm.DnsNodeInfoItem{
						WalletAddr:  item.WalletAddr.ToBase58(),
						IP:          string(item.IP),
						Port:        string(item.Port),
						InitDeposit: item.InitDeposit,
						PeerPubKey:  item.PeerPubKey,
					})
				}
				counter++
			}
			if counter == 0 {
				PrintInfoMsg("The total number of dns hosts info is 0")
			}
			return nil
		} else if dnsNIRsp.NodeInfoItem != nil {
			PrintInfoMsg("\nGet dns host info success.")
			PrintJsonObject(httpComm.DnsNodeInfoItem{
				WalletAddr:  dnsNIRsp.NodeInfoItem.WalletAddr.ToBase58(),
				IP:          string(dnsNIRsp.NodeInfoItem.IP),
				Port:        string(dnsNIRsp.NodeInfoItem.Port),
				InitDeposit: dnsNIRsp.NodeInfoItem.InitDeposit,
				PeerPubKey:  dnsNIRsp.NodeInfoItem.PeerPubKey,
			})
			return nil
		}
	} else {
		PrintInfoMsg("\nGet dns host info success. result is null")
		return nil
	}
	return nil
}

func checkTopology(ctx *cli.Context) error {
	return nil
}
