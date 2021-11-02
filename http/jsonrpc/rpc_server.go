/*
 * Copyright (C) 2019 The themis Authors
 * This file is part of The themis library.
 *
 * The themis is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The themis is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The themis.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package jsonrpc privides a function to start json rpc server
package jsonrpc

import (
	"net/http"
	"strconv"

	"fmt"

	cfg "github.com/saveio/scan/common/config"
	"github.com/saveio/scan/http/base/rpc"
	"github.com/saveio/themis/common/log"
)

const (
	SAVEDNS_DIR string = "/dns"
)

func StartRPCServer() error {
	log.Debug("startrpcserver")
	http.HandleFunc(SAVEDNS_DIR, rpc.Handle)

	rpc.HandleFunc("getbestblockhash", rpc.GetBestBlockHash)
	rpc.HandleFunc("getblock", rpc.GetBlock)
	rpc.HandleFunc("getblockcount", rpc.GetBlockCount)
	rpc.HandleFunc("getblockhash", rpc.GetBlockHash)
	rpc.HandleFunc("getconnectioncount", rpc.GetConnectionCount)
	//HandleFunc("getrawmempool", GetRawMemPool)

	rpc.HandleFunc("getrawtransaction", rpc.GetRawTransaction)
	rpc.HandleFunc("sendrawtransaction", rpc.SendRawTransaction)
	rpc.HandleFunc("getstorage", rpc.GetStorage)
	rpc.HandleFunc("getversion", rpc.GetNodeVersion)
	rpc.HandleFunc("getnetworkid", rpc.GetNetworkId)
	// rpc.HandleFunc("getstatemerkleroot", rpc.GetStateMerkleRoot)

	rpc.HandleFunc("getcontractstate", rpc.GetContractState)
	rpc.HandleFunc("getmempooltxcount", rpc.GetMemPoolTxCount)
	rpc.HandleFunc("getmempooltxstate", rpc.GetMemPoolTxState)
	rpc.HandleFunc("getsmartcodeevent", rpc.GetSmartCodeEvent)
	rpc.HandleFunc("getblockheightbytxhash", rpc.GetBlockHeightByTxHash)

	rpc.HandleFunc("getbalance", rpc.GetBalance)
	rpc.HandleFunc("getallowance", rpc.GetAllowance)
	rpc.HandleFunc("getmerkleproof", rpc.GetMerkleProof)
	rpc.HandleFunc("getblocktxsbyheight", rpc.GetBlockTxsByHeight)
	rpc.HandleFunc("getgasprice", rpc.GetGasPrice)

	rpc.HandleFunc("checktorrent", rpc.CheckTorrent)
	rpc.HandleFunc("regendpoint", rpc.EndPointReg)
	rpc.HandleFunc("updateendpoint", rpc.EndPointUpdate)
	rpc.HandleFunc("unRegendpoint", rpc.EndPointUnReg)
	rpc.HandleFunc("reqendpoint", rpc.EndPointReq)

	rpc.HandleFunc("registerdns", rpc.RegisterDns)
	rpc.HandleFunc("unregisterdns", rpc.UnregisterDns)
	rpc.HandleFunc("quitdns", rpc.QuitDns)
	rpc.HandleFunc("adddnspos", rpc.AddDnsPos)
	rpc.HandleFunc("reducednspos", rpc.ReduceDnsPos)
	rpc.HandleFunc("getregisterdnsinfo", rpc.GetRegisterDnsInfo)
	rpc.HandleFunc("getdnshostinfo", rpc.GetDnsHostInfo)

	rpc.HandleFunc("initprogress", rpc.GetChannelInitProgress)
	rpc.HandleFunc("joindnschannels", rpc.JoinDnsNodesChannels)
	rpc.HandleFunc("openchannel", rpc.OpenChannel)
	rpc.HandleFunc("closechannel", rpc.CloseChannel)
	rpc.HandleFunc("depositchannel", rpc.DepositToChannel)
	rpc.HandleFunc("transferchannel", rpc.TransferToSomebody)
	rpc.HandleFunc("mediatransferchannel", rpc.MediaTransferToSomebody)
	rpc.HandleFunc("withdrawchannel", rpc.WithdrawChannel)
	rpc.HandleFunc("getallchannels", rpc.GetAllChannels)
	rpc.HandleFunc("querychanneldeposit", rpc.QueryChannelDeposit)
	rpc.HandleFunc("queryhostinfo", rpc.QueryHostInfo)
	rpc.HandleFunc("getfee", rpc.GetFee)
	rpc.HandleFunc("setfee", rpc.SetFee)

	err := http.ListenAndServe(":"+strconv.Itoa(int(cfg.Parameters.Base.PortBase+cfg.Parameters.Base.JsonRpcPortOffset)), nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	log.Info("Rpc Listen at port: ", strconv.Itoa(int(cfg.Parameters.Base.PortBase+cfg.Parameters.Base.JsonRpcPortOffset)))
	return nil
}
