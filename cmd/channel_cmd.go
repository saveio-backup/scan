package cmd

import (
	"github.com/saveio/scan/cmd/flags"
	"github.com/saveio/scan/cmd/utils"

	//"github.com/saveio/scan/config"
	"github.com/urfave/cli"
)

var ChannelCommand = cli.Command{
	Action:      cli.ShowSubcommandHelp,
	Name:        "channel",
	Usage:       "Manage state channels",
	Description: "Manage state channels",
	Subcommands: []cli.Command{
		{
			Action:      initProgress,
			Name:        "initprogress",
			Usage:       "Get channel init progress",
			ArgsUsage:   " ",
			Flags:       []cli.Flag{},
			Description: "Get channel init progress",
		},
		{
			Action:    openChannel,
			Name:      "open",
			Usage:     "Open a payment channel",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
			},
			Description: "Open a payment channel with partner",
		},
		// {
		// 	Action:    closeChannel,
		// 	Name:      "close",
		// 	Usage:     "Close a payment channel",
		// 	ArgsUsage: " ",
		// 	Flags: []cli.Flag{
		// 		flags.PartnerAddressFlag,
		// 	},
		// 	Description: "Close a payment channel with partner",
		// },
		{
			Action:    depositToChannel,
			Name:      "deposit",
			Usage:     "Deposit token to channel",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
				flags.TotalDepositFlag,
			},
			Description: "Deposit token to channel with specified partner",
		},
		{
			Action:    withdrawChannel,
			Name:      "withdraw",
			Usage:     "Withdraw channel deposit",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
				flags.AmountFlag,
			},
			Description: "Withdraw deposit of channel which belong to owner and partner",
		},
		{
			Action:    transferToSomebody,
			Name:      "transfer",
			Usage:     "Make payment through channel",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.TargetAddressFlag,
				flags.AmountFlag,
				flags.PaymentIDFlag,
			},
			Description: "Transfer some token from owner to target with specified payment ID",
		},
		{
			Action:      getAllChannels,
			Name:        "list",
			Usage:       "Show all channels",
			ArgsUsage:   " ",
			Flags:       []cli.Flag{},
			Description: "Show all channels info which belong to current owner",
		},
		// {
		// 	Action:    getCurrentBalance,
		// 	Name:      "balance",
		// 	Usage:     "Get current balance",
		// 	ArgsUsage: " ",
		// 	Flags: []cli.Flag{
		// 		flags.PartnerAddressFlag,
		// 	},
		// 	Description: "Get current channel balance which belong to current owner",
		// },
		{
			Action:    queryChannelDeposit,
			Name:      "querydeposit",
			Usage:     "Query channel deposit",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
			},
			Description: "Query deposit of channel which belong to owner and partner",
		},
		{
			Action:    queryHostInfo,
			Name:      "queryhost",
			Usage:     "Query host info",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
			},
			Description: "Query host info of partner",
		},
		{
			Action:    cooperativeSettle,
			Name:      "cooperativeSettle",
			Usage:     "Cooperative settle",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
			},
			Description: "settle cooperatively of channel which belong to owner and partner",
		},
	},
}

func initProgress(ctx *cli.Context) error {
	SetRpcPort(ctx)
	progress, failed := utils.CheckChannelInitProgress()
	if failed != nil {
		PrintErrorMsg("\nCheck channel init progress failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}
	PrintInfoMsg("\nCheck chennal init preogress success. Progress:\n %+v", progress)
	return nil
}

func openChannel(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))

	chanRsp, failed := utils.OpenChannel(partnerAddr)
	if failed != nil {
		PrintErrorMsg("\nOpen channel failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}
	PrintInfoMsg("\nOpen channel success. Channel ID: %d", chanRsp.Id)
	return nil
}

// func closeChannel(ctx *cli.Context) error {
// 	SetRpcPort(ctx)

// 	if ctx.NumFlags() < 1 {
// 		PrintErrorMsg("Missing argument.")
// 		cli.ShowSubcommandHelp(ctx)
// 		return nil
// 	}

// 	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))

// 	_, failed := utils.CloseChannel(partnerAddr)
// 	if failed != nil {
// 		PrintErrorMsg("\nClose channel failed. Failed message:")
// 		PrintJsonObject(failed)
// 		return nil
// 	}
// 	PrintInfoMsg("\nClose channel success.")
// 	return nil
// }

func depositToChannel(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))
	totalDeposit := ctx.Uint64(flags.GetFlagName(flags.TotalDepositFlag))

	_, failed := utils.DepositToChannel(partnerAddr, totalDeposit)
	if failed != nil {
		PrintErrorMsg("\nDeposit to channel failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}
	PrintInfoMsg("\nDeposit to channel success.")
	return nil
}

func withdrawChannel(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))
	amount := ctx.Uint64(flags.GetFlagName(flags.AmountFlag))

	_, failed := utils.WithdrawChannel(partnerAddr, amount)
	if failed != nil {
		PrintErrorMsg("\nWithdraw channel failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}
	PrintInfoMsg("\nWithdraw channel success.")
	return nil
}

func transferToSomebody(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 3 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	partnerAddr := ctx.String(flags.GetFlagName(flags.TargetAddressFlag))
	amount := ctx.Uint64(flags.GetFlagName(flags.AmountFlag))
	paymentId := ctx.Uint(flags.GetFlagName(flags.PaymentIDFlag))

	_, failed := utils.TransferToSomebody(partnerAddr, amount, paymentId)
	if failed != nil {
		PrintErrorMsg("\nTransfer by channel failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}
	PrintInfoMsg("\nTransfer by channel success.")
	return nil
}

func getAllChannels(ctx *cli.Context) error {
	SetRpcPort(ctx)

	channelInfos, failed := utils.GetAllChannels()
	if failed != nil {
		PrintErrorMsg("\nShow all channels info failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	} else if channelInfos != nil {
		PrintInfoMsg("\nShow all channels info success. Default limit 50:")
		PrintInfoMsg("The total number of channels is %d", len(channelInfos.Channels))
		PrintInfoMsg("The total balance of channels is %d, and BalanceFormat: %s", channelInfos.Balance, channelInfos.BalanceFormat)
		counter := 0
		for _, item := range channelInfos.Channels {
			if counter < 50 {
				PrintInfoMsg("Index %d:", counter)
				PrintJsonObject(item)
			}
			counter++
		}
	}

	return nil
}

// func getCurrentBalance(ctx *cli.Context) error {
// 	if ctx.NumFlags() < 1 {
// 		PrintErrorMsg("Missing argument.")
// 		cli.ShowSubcommandHelp(ctx)
// 		return nil
// 	}

// 	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))
// 	currBalanceRsp, failed := utils.GetCurrentBalance(partnerAddr)
// 	if failed != nil {
// 		PrintErrorMsg("\nGet channel current balance failed. Failed message:")
// 		PrintJsonObject(failed)
// 		return nil
// 	}
// 	PrintInfoMsg("\nGet channel current balance success. TotalDepositBalance msg:")
// 	PrintJsonObject(currBalanceRsp)
// 	return nil
// }

func queryChannelDeposit(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))
	totalDepositBalanceRsp, failed := utils.QueryChannelDeposit(partnerAddr)
	if failed != nil {
		PrintErrorMsg("\nQuery channel deposit balance failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}
	PrintInfoMsg("\nQuery channel deposit balance success. TotalDepositBalance msg:")
	PrintJsonObject(totalDepositBalanceRsp)
	return nil
}

func cooperativeSettle(ctx *cli.Context) error {
	return nil
}

func queryHostInfo(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))
	chanHostRsp, failed := utils.QueryHostInfo(partnerAddr)
	if failed != nil {
		PrintErrorMsg("\nTransfer by channel failed. Failed message:")
		PrintJsonObject(failed)
		return nil
	}
	PrintInfoMsg("\nTransfer by channel success. Channel host info msg:")
	PrintJsonObject(chanHostRsp)
	return nil
}
