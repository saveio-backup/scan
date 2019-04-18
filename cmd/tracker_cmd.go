/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-15 
*/
package cmd
import (
	"github.com/urfave/cli"
	"github.com/oniio/oniDNS/cmd/common"
	"github.com/oniio/oniDNS/cmd/utils"
)

var EndPointCommand=cli.Command{
	Action:cli.ShowSubcommandHelp,
	Name:"endpoint",
	Usage:"Manage endpoint(reg|unReg|req|update)",
	Description:"Local DDNS node (reg|unReg|req|update)",
	Subcommands: []cli.Command{
		{
			Action:regEndPoint,
			Name:"reg",
			Usage:"Reg wallet address and host",
			ArgsUsage: "[sub-command options]",
			Flags: []cli.Flag{
				utils.WalletFlag,
				utils.HostFlag,
			},
		},


	},

}

func regEndPoint(ctx *cli.Context)error{
	SetRpcPort(ctx)
	//if ctx.NumFlags() < 2 {
	//	PrintErrorMsg("Missing argument.")
	//	cli.ShowSubcommandHelp(ctx)
	//	return nil
	//}

	client, err := common.OpenWallet(ctx)
	if err!=nil{
		return err
	}
	pw,err:=common.GetPasswd(ctx)
	if err!=nil{
		PrintErrorMsg("regEndPoint GetPasswd error:%s\n",err)
		return err
	}
	acc,err:=client.GetDefaultAccount(pw)
	if err!=nil{
		PrintErrorMsg("regEndPoint GetDefaultAccount error:%s\n",err)
		return err
	}
	a:=acc.Address
	wAddr:=a.ToHexString()
	host:=ctx.String(utils.GetFlagName(utils.HostFlag))
	//endPoint:=httpComm.EndPointRsp{
	//	Wallet:wAddr,
	//	Host:host,
	//}
	//em,err:=json.Marshal(endPoint)
	//if err!=nil{
	//	PrintErrorMsg("regEndPoint json.Marshal error:%s\n",err)
	//}
	//sigdata,err:=utils.Sign(em,acc)
	if err!=nil{
		PrintErrorMsg("regEndPoint utils.Sign error:%s\n",err)
	}
	err=utils.RegEndPoint(wAddr,host)
	if err!=nil{
		return err
		PrintErrorMsg("Register wallet:%s,host:%s",wAddr,host)
	}
	return nil
}

