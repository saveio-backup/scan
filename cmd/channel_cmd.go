package cmd

import (
	"github.com/saveio/scan/cmd/flags"
	"github.com/saveio/scan/cmd/utils"
	"github.com/saveio/themis/common/constants"
	"math"
	"strconv"

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
			Name:        "progress",
			Usage:       "Get channel init progress",
			ArgsUsage:   " ",
			Flags:       []cli.Flag{},
			Description: "Get channel init progress",
		},
		{
			Action:      join,
			Name:        "join",
			Usage:       "Join dns nodes channels",
			ArgsUsage:   " ",
			Flags:       []cli.Flag{},
			Description: "Join dns nodes channels",
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
		{
			Action:    closeChannel,
			Name:      "close",
			Usage:     "Close a payment channel",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
			},
			Description: "Close a payment channel with partner",
		},
		{
			Action:    depositToChannel,
			Name:      "deposit",
			Usage:     "Deposit token to channel",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
				flags.DepositFlag,
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
			Action:    mediaTransferToSomebody,
			Name:      "mediaTransfer",
			Usage:     "Make payment through the channel with a media node",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PaymentIDFlag,
				flags.AmountFlag,
				flags.MediaAddressFlag,
				flags.TargetAddressFlag,
			},
			Description: "Transfer some token from owner to target with specified payment ID through media node",
		},
		{
			Action:    getAllChannels,
			Name:      "list",
			Usage:     "Show channels by paging, or target by address",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PageFlag,
				flags.PartnerAddressFlag,
			},
			Description: "Show channels by paging, or target by address",
		},
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
		{
			Action:    getFee,
			Name:      "getfee",
			Usage:     "Get fee schedule in mediation",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.ChannelIdFlag,
			},
			Description: "Query fee schedule in mediation",
		},
		{
			Action:    setFee,
			Name:      "setfee",
			Usage:     "Setup fee schedule in mediation",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.FlatFlag,
				flags.ProportionalFlag,
			},
			Description: "Setup fee schedule in mediation",
		},
		{
			Action:    getPenalty,
			Name:      "getPenalty",
			Usage:     "Get penalty setup",
			ArgsUsage: " ",
			Flags: []cli.Flag{
			},
			Description: "Get penalty setup",
		},
		{
			Action:    setPenalty,
			Name:      "setPenalty",
			Usage:     "Set penalty setup",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.FeePenaltyFlag,
				flags.DiversityPenaltyFlag,
			},
			Description: "Set penalty setup",
		},
	},
}

func initProgress(ctx *cli.Context) error {
	SetRpcPort(ctx)
	progress, failed := utils.CheckChannelInitProgress()
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}
	PrintInfoMsg("\nCheck chennal init preogress success. Progress:\n %+v", progress)
	return nil
}

func join(ctx *cli.Context) error {
	SetRpcPort(ctx)
	failed := utils.JoinDnsChannels()
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}
	PrintInfoMsg("\nJoin dns nodes channels success.\n")
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
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}
	PrintInfoMsg("\nOpen channel success. Channel ID: %d", chanRsp.Id)
	return nil
}

func closeChannel(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))

	_, failed := utils.CloseChannel(partnerAddr)
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}
	PrintInfoMsg("\nClose channel success.")
	return nil
}

func depositToChannel(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 3 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))
	taStr := ctx.String(flags.GetFlagName(flags.DepositFlag))

	ta, err := strconv.ParseFloat(taStr, 10)
	if err != nil || ta <= 0 {
		PrintErrorMsg("Deposit amount must larger than 0")
		return nil
	}
	totalDeposit := uint64(ta * math.Pow10(constants.USDT_DECIMALS))

	_, failed := utils.DepositToChannel(partnerAddr, totalDeposit)
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
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
	amountStr := ctx.String(flags.GetFlagName(flags.AmountFlag))

	amount, err := strconv.ParseFloat(amountStr, 10)
	if err != nil || amount <= 0 {
		PrintErrorMsg("Withdraw amount must larger than 0")
		return nil
	}

	realAmount := uint64(amount * math.Pow10(constants.USDT_DECIMALS))
	_, failed := utils.WithdrawChannel(partnerAddr, realAmount)
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
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
	amountStr := ctx.String(flags.GetFlagName(flags.AmountFlag))
	paymentId := ctx.Uint(flags.GetFlagName(flags.PaymentIDFlag))

	amount, err := strconv.ParseFloat(amountStr, 10)
	if err != nil || amount < 0 {
		return nil
	}
	realAmount := uint64(amount * math.Pow10(constants.USDT_DECIMALS))

	_, failed := utils.TransferToSomebody(partnerAddr, realAmount, paymentId)
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}
	PrintInfoMsg("\nTransfer by channel success.")
	return nil
}

func mediaTransferToSomebody(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NumFlags() < 3 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	mediaAddr := ctx.String(flags.GetFlagName(flags.MediaAddressFlag))
	partnerAddr := ctx.String(flags.GetFlagName(flags.TargetAddressFlag))
	amountStr := ctx.String(flags.GetFlagName(flags.AmountFlag))
	paymentId := ctx.Uint(flags.GetFlagName(flags.PaymentIDFlag))

	amount, err := strconv.ParseFloat(amountStr, 10)
	if err != nil || amount < 0 {
		return nil
	}
	realAmount := uint64(amount * math.Pow10(constants.USDT_DECIMALS))

	_, failed := utils.MediaTransferToSomebody(paymentId, realAmount, mediaAddr, partnerAddr)
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}
	PrintInfoMsg("\nMediaTransferToSomebody by channel success.")
	return nil
}

func getAllChannels(ctx *cli.Context) error {
	SetRpcPort(ctx)

	page := ctx.Int(flags.GetFlagName(flags.PageFlag))
	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))

	channelInfos, failed := utils.GetAllChannels()
	if failed != nil {
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	} else if channelInfos != nil {
		if partnerAddr != "" {
			exist := false
			for _, item := range channelInfos.Channels {
				if item.Address == partnerAddr {
					PrintJsonObject(item)
					exist = true
					break
				}
			}
			if !exist {
				PrintInfoMsg("do not exist.")
			}
		} else {
			PrintInfoMsg("Show all channels limit 50 in default, can be paging used by flag [--page]")
			PrintInfoMsg("The total number of channels is %d", len(channelInfos.Channels))
			PrintInfoMsg("The total balance of channels is %d, and BalanceFormat: %s", channelInfos.Balance, channelInfos.BalanceFormat)
			counter, pagePrev, pageNext := 0, 50*page, 50*(page+1)
			for _, item := range channelInfos.Channels {
				if pagePrev <= counter && counter < pageNext {
					PrintInfoMsg("Index %d:", counter)
					PrintJsonObject(item)
				}
				counter++
			}
		}
	}

	return nil
}

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
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}
	PrintJsonObject(totalDepositBalanceRsp)
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
		PrintErrorMsg("%v\n", failed.FailedMsg)
		return nil
	}
	PrintJsonObject(chanHostRsp)
	return nil
}

func cooperativeSettle(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	addr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))

	failMsg := utils.CooperativeSettle(addr)
	if failMsg != nil {
		PrintErrorMsg("%v\n", failMsg.FailedMsg)
		return nil
	}
	PrintInfoMsg("\nCooperativeSettlee success.")
	return nil
}

func getFee(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 0 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	channelID := ctx.Uint64(flags.GetFlagName(flags.ChannelIdFlag))

	fee, failedRsp := utils.GetFee(channelID)
	if failedRsp != nil {
		PrintErrorMsg("%v\n", failedRsp.FailedMsg)
		return nil
	}
	PrintJsonObject(fee)
	return nil
}

func setFee(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	flatStr := ctx.String(flags.GetFlagName(flags.FlatFlag))
	proStr := ctx.String(flags.GetFlagName(flags.ProportionalFlag))

	flat, err := strconv.ParseFloat(flatStr, 10)
	if err != nil || flat < 0 {
		return nil
	}
	realFlat := uint64(flat * math.Pow10(constants.USDT_DECIMALS))

	pro, err := strconv.ParseFloat(proStr, 10)
	if err != nil || pro < 0 {
		return nil
	}
	realPro := uint64(pro * math.Pow10(constants.USDT_DECIMALS))

	failMsg := utils.SetFee(realFlat, realPro)
	if failMsg != nil {
		PrintErrorMsg("%v\n", failMsg.FailedMsg)
		return nil
	}
	PrintInfoMsg("\nSetup fee schedule success.")
	return nil
}


func getPenalty(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 0 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	res, failedRsp := utils.GetPenalty()
	if failedRsp != nil {
		PrintErrorMsg("%v\n", failedRsp.FailedMsg)
		return nil
	}
	PrintJsonObject(res)
	return nil
}

func setPenalty(ctx *cli.Context) error {
	SetRpcPort(ctx)

	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	fp := ctx.Float64(flags.GetFlagName(flags.FeePenaltyFlag))
	dp := ctx.Float64(flags.GetFlagName(flags.DiversityPenaltyFlag))

	failedRsp := utils.SetPenalty(fp, dp)
	if failedRsp != nil {
		PrintErrorMsg("%v\n", failedRsp.FailedMsg)
		return nil
	}
	return nil
}
