package cmd

import (
	"github.com/anacrolix/dht/krpc"
	"github.com/saveio/scan/cmd/common"
	"github.com/saveio/scan/cmd/flags"
	"github.com/saveio/scan/cmd/utils"

	//"github.com/saveio/scan/config"

	"github.com/urfave/cli"
)

var TrackerCommand = cli.Command{
	Action:      cli.ShowSubcommandHelp,
	Name:        "tracker",
	Usage:       "Maintain torrent and endpoint map",
	Description: "Maintain torrent and endpoint map",
	Subcommands: []cli.Command{
		{
			Action:      cli.ShowSubcommandHelp,
			Name:        "torrent",
			Usage:       "Manage torrents",
			Description: "Manage torrents",
			Subcommands: []cli.Command{
				{
					Action:    checkTorrent,
					Name:      "check",
					Usage:     "Check map of hash and torrent",
					ArgsUsage: "[sub-command options]",
					Flags: []cli.Flag{
						flags.FileHashFlag,
					},
				},
			},
		},
		{
			Action:      cli.ShowSubcommandHelp,
			Name:        "endpoint",
			Usage:       "Manage endpoints",
			Description: "Manage endpoints",
			Subcommands: []cli.Command{
				{
					Action:    regEndPoint,
					Name:      "reg",
					Usage:     "Reg wallet address and host",
					ArgsUsage: "[sub-command options]",
					Flags: []cli.Flag{
						flags.WalletFlag,
						flags.HostFlag,
					},
				},
				{
					Action:    updateEndPoint,
					Name:      "update",
					Usage:     "Update the host of your wallet address",
					ArgsUsage: "[sub-command options]",
					Flags: []cli.Flag{
						flags.WalletFlag,
						flags.HostFlag,
					},
				},
				{
					Action:    unRegEndPoint,
					Name:      "unreg",
					Usage:     "UnReg wallet address and host",
					ArgsUsage: "[sub-command options]",
					Flags: []cli.Flag{
						flags.WalletFlag,
					},
				},
				{
					Action:    reqEndPoint,
					Name:      "req",
					Usage:     "query the host of your wallet address or others",
					ArgsUsage: "[sub-command options]",
					Flags: []cli.Flag{
						flags.WalletFlag,
					},
				},
			},
		},
	},
}

func checkTorrent(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		return cli.ShowSubcommandHelp(ctx)
	}

	fileHash := ctx.String(flags.GetFlagName(flags.FileHashFlag))

	peers, failed := utils.CheckTorrent(fileHash)
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}

	if len(peers.Peers) == 0 {
		PrintInfoMsg("\nCheck torrent success. no peers")
	} else {
		PrintInfoMsg("\nCheck torrent success. peers:")
		for _, peer := range peers.Peers {

			PrintInfoMsg("%s\t%v", krpc.NodeAddr{IP: peer.IP[:], Port: int(peer.Port)}.String(), peer.Timestamp.Format("2006-01-02 15:04:05"))
		}
	}
	return nil
}

func regEndPoint(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		return cli.ShowSubcommandHelp(ctx)
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

	wAddr = ctx.String(flags.GetFlagName(flags.WalletFlag))
	host := ctx.String(flags.GetFlagName(flags.HostFlag))

	endpoint, failed := utils.RegEndPoint(wAddr, host)
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}
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

	wAddr = ctx.String(flags.GetFlagName(flags.WalletFlag))
	host := ctx.String(flags.GetFlagName(flags.HostFlag))

	endpoint, failed := utils.UpdateEndPoint(wAddr, host)
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}
	PrintInfoMsg("Update endpoint success.")
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

	wAddr = ctx.String(flags.GetFlagName(flags.WalletFlag))

	endpoint, failed := utils.UnRegEndPoint(wAddr)
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}

	PrintInfoMsg("Unregister endpoint success.")
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

	wAddr = ctx.String(flags.GetFlagName(flags.WalletFlag))

	endpoint, failed := utils.ReqEndPoint(wAddr)
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}
	PrintJsonObject(endpoint)
	return nil
}
