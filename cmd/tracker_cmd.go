/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-15
 */
package cmd

import (
	"github.com/oniio/oniDNS/cmd/common"
	"github.com/oniio/oniDNS/cmd/utils"
	//"github.com/oniio/oniDNS/config"
	"github.com/urfave/cli"
)

var EndPointCommand = cli.Command{
	Action:      cli.ShowSubcommandHelp,
	Name:        "endpoint",
	Usage:       "Manage endpoint(reg|unReg|req|update)",
	Description: "Local endpoint (reg|unReg|req|update)",
	Subcommands: []cli.Command{
		{
			Action:    regEndPoint,
			Name:      "reg",
			Usage:     "Reg wallet address and host",
			ArgsUsage: "[sub-command options]",
			Flags: []cli.Flag{
				utils.WalletFlag,
				utils.HostFlag,
			},
		},
		{
			Action:    updateEndPoint,
			Name:      "update",
			Usage:     "Update the host of your wallet address",
			ArgsUsage: "[sub-command options]",
			Flags: []cli.Flag{
				utils.HostFlag,
			},
		},
		{
			Action:    unRegEndPoint,
			Name:      "unreg",
			Usage:     "UnReg wallet address and host",
			ArgsUsage: "[sub-command options]",
		},
		{
			Action:    reqEndPoint,
			Name:      "req",
			Usage:     "query the host of your wallet address or others",
			ArgsUsage: "[sub-command options]",
			Flags: []cli.Flag{
				utils.WalletFlag,
			},
		},
	},
}

func regEndPoint(ctx *cli.Context) error {
	SetRpcPort(ctx)
	//if ctx.NumFlags() < 2 {
	//	PrintErrorMsg("Missing argument.")
	//	cli.ShowSubcommandHelp(ctx)
	//	return nil
	//}
	var wAddr string
	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("regEndPoint GetPasswd error:%s\n", err)
		return err
	}
	acc, err := client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("regEndPoint GetDefaultAccount error:%s\n", err)
		return err
	}
	a := acc.Address
	if ctx.IsSet(utils.GetFlagName(utils.WalletFlag)) {
		wAddr = ctx.String(utils.GetFlagName(utils.WalletFlag))
	} else {
		wAddr = a.ToBase58()
	}
	host := ctx.String(utils.GetFlagName(utils.HostFlag))
	if err != nil {
		PrintErrorMsg("regEndPoint utils.Sign error:%s\n", err)
	}
	err = utils.RegEndPoint(wAddr, host)
	if err != nil {
		return err
		PrintErrorMsg("Register wallet:%s,host:%s", wAddr, host)
	}
	return nil
}

func updateEndPoint(ctx *cli.Context) error {
	SetRpcPort(ctx)
	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("regEndPoint GetPasswd error:%s\n", err)
		return err
	}
	acc, err := client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("regEndPoint GetDefaultAccount error:%s\n", err)
		return err
	}
	a := acc.Address
	wAddr := a.ToBase58()
	host := ctx.String(utils.GetFlagName(utils.HostFlag))
	if err != nil {
		PrintErrorMsg("regEndPoint utils.Sign error:%s\n", err)
	}
	err = utils.UpdateEndPoint(wAddr, host)
	if err != nil {
		return err
		PrintErrorMsg("Register wallet:%s,host:%s", wAddr, host)
	}
	return nil
}

func unRegEndPoint(ctx *cli.Context) error {
	SetRpcPort(ctx)
	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("regEndPoint GetPasswd error:%s\n", err)
		return err
	}
	acc, err := client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("regEndPoint GetDefaultAccount error:%s\n", err)
		return err
	}
	a := acc.Address
	wAddr := a.ToBase58()
	err = utils.UnRegEndPoint(wAddr)
	if err != nil {
		return err
		PrintErrorMsg("unRegister wallet:%s", wAddr)
	}
	return nil
}

func reqEndPoint(ctx *cli.Context) error {
	SetRpcPort(ctx)
	var wAddr string
	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("regEndPoint GetPasswd error:%s\n", err)
		return err
	}
	acc, err := client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("regEndPoint GetDefaultAccount error:%s\n", err)
		return err
	}
	a := acc.Address
	if ctx.IsSet(utils.GetFlagName(utils.WalletFlag)) {
		wAddr = ctx.String(utils.GetFlagName(utils.WalletFlag))
	} else {
		wAddr = a.ToBase58()
	}
	host := ctx.String(utils.GetFlagName(utils.WalletFlag))
	if err != nil {
		PrintErrorMsg("regEndPoint utils.Sign error:%s\n", err)
	}
	err = utils.ReqEndPoint(wAddr)
	if err != nil {
		return err
		PrintErrorMsg("Register wallet:%s,host:%s", wAddr, host)
	}
	return nil
}
