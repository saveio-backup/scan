/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-15
 */
package cmd

import (
	"github.com/saveio/scan/cmd/common"
	"github.com/saveio/scan/cmd/utils"

	//"github.com/saveio/scan/config"

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
				utils.WalletFlag,
				utils.HostFlag,
			},
		},
		{
			Action:    unRegEndPoint,
			Name:      "unreg",
			Usage:     "UnReg wallet address and host",
			ArgsUsage: "[sub-command options]",
			Flags: []cli.Flag{
				utils.WalletFlag,
			},
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

	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	var wAddr string
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

	wAddr = ctx.String(utils.GetFlagName(utils.WalletFlag))
	host := ctx.String(utils.GetFlagName(utils.HostFlag))

	endpoint, failed := utils.RegEndPoint(wAddr, host)
	if failed != nil {
		PrintErrorMsg("Register endpoint failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}

	PrintInfoMsg("Register endpoint success. EndPoint message:")
	PrintJsonObject(endpoint)
	return nil
}

func updateEndPoint(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	var wAddr string
	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("GetPasswd error:%s\n", err)
		return err
	}
	_, err = client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("GetDefaultAccount from wallet and password error, ErrMsg:")
		return err
	}

	wAddr = ctx.String(utils.GetFlagName(utils.WalletFlag))
	host := ctx.String(utils.GetFlagName(utils.HostFlag))

	endpoint, failed := utils.UpdateEndPoint(wAddr, host)
	if failed != nil {
		PrintErrorMsg("Update endpoint failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}

	PrintInfoMsg("Update endpoint success. EndPoint updated message:")
	PrintJsonObject(endpoint)
	return nil
}

func unRegEndPoint(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	var wAddr string
	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	pw, err := common.GetPasswd(ctx)
	if err != nil {
		PrintErrorMsg("GetDefaultAccount from wallet and password error, ErrMsg:")
		return err
	}
	_, err = client.GetDefaultAccount(pw)
	if err != nil {
		PrintErrorMsg("regEndPoint GetDefaultAccount error:%s\n", err)
		return err
	}

	wAddr = ctx.String(utils.GetFlagName(utils.WalletFlag))

	endpoint, failed := utils.UnRegEndPoint(wAddr)
	if failed != nil {
		PrintErrorMsg("Unregister endpoint failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}

	PrintInfoMsg("Unregister endpoint success. Endpoint unregister message:")
	PrintJsonObject(endpoint)
	return nil
}

func reqEndPoint(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	var wAddr string
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

	wAddr = ctx.String(utils.GetFlagName(utils.WalletFlag))

	endpoint, failed := utils.ReqEndPoint(wAddr)
	if failed != nil {
		PrintErrorMsg("Request endpoint failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}

	PrintInfoMsg("Request endpoint success. EndPoint message:")
	PrintJsonObject(endpoint)
	return nil
}
