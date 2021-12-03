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

package rpc

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/saveio/dsp-go-sdk/consts"
	chanCom "github.com/saveio/pylons/common"
	"github.com/saveio/pylons/transfer"
	httpComm "github.com/saveio/scan/http/base/common"
	berr "github.com/saveio/scan/http/base/error"
	"github.com/saveio/scan/service"
	"github.com/saveio/scan/tracker"
	"github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/payload"
	scom "github.com/saveio/themis/core/store/common"
	"github.com/saveio/themis/core/types"
	ontErrors "github.com/saveio/themis/errors"
	bactor "github.com/saveio/themis/http/base/actor"
	bcomn "github.com/saveio/themis/http/base/common"
)

//get best block hash
func GetBestBlockHash(params []interface{}) map[string]interface{} {
	hash := bactor.CurrentBlockHash()
	return responseSuccess(hash.ToHexString())
}

// get block by height or hash
// Input JSON string examples for getblock method as following:
//   {"jsonrpc": "2.0", "method": "getblock", "params": [1], "id": 0}
//   {"jsonrpc": "2.0", "method": "getblock", "params": ["aabbcc.."], "id": 0}
func GetBlock(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	var err error
	var hash common.Uint256
	switch (params[0]).(type) {
	// block height
	case float64:
		index := uint32(params[0].(float64))
		hash = bactor.GetBlockHashFromStore(index)
		if hash == common.UINT256_EMPTY {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		// block hash
	case string:
		str := params[0].(string)
		hash, err = common.Uint256FromHexString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	block, err := bactor.GetBlockFromStore(hash)
	if err != nil {
		return responsePack(berr.UNKNOWN_BLOCK, "unknown block")
	}
	if len(params) >= 2 {
		switch (params[1]).(type) {
		case float64:
			json := uint32(params[1].(float64))
			if json == 1 {
				return responseSuccess(bcomn.GetBlockInfo(block))
			}
		default:
			return responsePack(berr.INVALID_PARAMS, "")
		}
	}
	return responseSuccess(common.ToHexString(block.ToArray()))
}

//get block height
func GetBlockCount(params []interface{}) map[string]interface{} {
	height := bactor.GetCurrentBlockHeight()
	return responseSuccess(height + 1)
}

//get block hash
// A JSON example for getblockhash method as following:
//   {"jsonrpc": "2.0", "method": "getblockhash", "params": [1], "id": 0}
func GetBlockHash(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	switch params[0].(type) {
	case float64:
		height := uint32(params[0].(float64))
		hash := bactor.GetBlockHashFromStore(height)
		if hash == common.UINT256_EMPTY {
			return responsePack(berr.UNKNOWN_BLOCK, "")
		}
		return responseSuccess(hash.ToHexString())
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
}

//get node connection count
func GetConnectionCount(params []interface{}) map[string]interface{} {
	count := bactor.GetConnectionCnt()
	return responseSuccess(count)
}

func GetRawMemPool(params []interface{}) map[string]interface{} {
	txs := []*bcomn.Transactions{}
	txpool := bactor.GetTxsFromPool(false)
	for _, t := range txpool {
		txs = append(txs, bcomn.TransArryByteToHexString(t))
	}
	if len(txs) == 0 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	return responseSuccess(txs)
}

//get memory pool transaction count
func GetMemPoolTxCount(params []interface{}) map[string]interface{} {
	count, err := bactor.GetTxnCount()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, nil)
	}
	return responseSuccess(count)
}

//get memory pool transaction state
func GetMemPoolTxState(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hash, err := common.Uint256FromHexString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		txEntry, err := bactor.GetTxFromPool(hash)
		if err != nil {
			return responsePack(berr.UNKNOWN_TRANSACTION, "unknown transaction")
		}
		attrs := []bcomn.TXNAttrInfo{}
		for _, t := range txEntry.Attrs {
			attrs = append(attrs, bcomn.TXNAttrInfo{t.Height, int(t.Type), int(t.ErrCode)})
		}
		info := bcomn.TXNEntryInfo{attrs}
		return responseSuccess(info)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
}

// get raw transaction in raw or json
// A JSON example for getrawtransaction method as following:
//   {"jsonrpc": "2.0", "method": "getrawtransaction", "params": ["transactioin hash in hex"], "id": 0}
func GetRawTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	var tx *types.Transaction
	var height uint32
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hash, err := common.Uint256FromHexString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		h, t, err := bactor.GetTxnWithHeightByTxHash(hash)
		if err != nil {
			return responsePack(berr.UNKNOWN_TRANSACTION, "unknown transaction")
		}
		height = h
		tx = t
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	if len(params) >= 2 {
		switch (params[1]).(type) {
		case float64:
			json := uint32(params[1].(float64))
			if json == 1 {
				txinfo := bcomn.TransArryByteToHexString(tx)
				txinfo.Height = height
				return responseSuccess(txinfo)
			}
		default:
			return responsePack(berr.INVALID_PARAMS, "")
		}
	}

	return responseSuccess(common.ToHexString(common.SerializeToBytes(tx)))
}

//get storage from contract
//   {"jsonrpc": "2.0", "method": "getstorage", "params": ["code hash", "key"], "id": 0}
func GetStorage(params []interface{}) map[string]interface{} {
	if len(params) < 2 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}

	var address common.Address
	var key []byte
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		var err error
		address, err = bcomn.GetAddress(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch params[1].(type) {
	case string:
		str := params[1].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		key = hex
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	value, err := bactor.GetStorageItem(address, key)
	if err != nil {
		if err == scom.ErrNotFound {
			return responseSuccess(nil)
		}
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responseSuccess(common.ToHexString(value))
}

//send raw transaction
// A JSON example for sendrawtransaction method as following:
//   {"jsonrpc": "2.0", "method": "sendrawtransaction", "params": ["raw transactioin in hex"], "id": 0}
func SendRawTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	var hash common.Uint256
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		raw, err := common.HexToBytes(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		txn, err := types.TransactionFromRawBytes(raw)
		if err != nil {
			return responsePack(berr.INVALID_TRANSACTION, "")
		}
		hash = txn.Hash()
		log.Debugf("SendRawTransaction recv %s", hash.ToHexString())
		if txn.TxType == types.InvokeNeo || txn.TxType == types.Deploy || txn.TxType == types.InvokeWasm {
			if len(params) > 1 {
				preExec, ok := params[1].(float64)
				if ok && preExec == 1 {
					result, err := bactor.PreExecuteContract(txn)
					if err != nil {
						log.Infof("PreExec: ", err)
						return responsePack(berr.SMARTCODE_ERROR, err.Error())
					}
					return responseSuccess(bcomn.ConvertPreExecuteResult(result))
				}
			}
		}

		log.Debugf("SendRawTransaction send to txpool %s", hash.ToHexString())
		if errCode, desc := bcomn.SendTxToPool(txn); errCode != ontErrors.ErrNoError {
			log.Warnf("SendRawTransaction verified %s error: %s", hash.ToHexString(), desc)
			return responsePack(int64(errCode), desc)
		}
		log.Debugf("SendRawTransaction verified %s", hash.ToHexString())
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responseSuccess(hash.ToHexString())
}

//get node version
func GetNodeVersion(params []interface{}) map[string]interface{} {
	return responseSuccess(config.Version)
}

// get networkid
func GetNetworkId(params []interface{}) map[string]interface{} {
	return responseSuccess(config.DefConfig.P2PNode.NetworkId)
}

//get contract state
func GetContractState(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	var contract *payload.DeployCode
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		address, err := bcomn.GetAddress(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		c, err := bactor.GetContractStateFromStore(address)
		if err != nil {
			return responsePack(berr.UNKNOWN_CONTRACT, "unknow contract")
		}
		contract = c
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	if len(params) >= 2 {
		switch (params[1]).(type) {
		case float64:
			json := uint32(params[1].(float64))
			if json == 1 {
				return responseSuccess(bcomn.TransPayloadToHex(contract))
			}
		default:
			return responsePack(berr.INVALID_PARAMS, "")
		}
	}
	w := bytes.NewBuffer(nil)
	contract.Serialize(w)
	return responseSuccess(common.ToHexString(w.Bytes()))
}

//get smartconstract event
func GetSmartCodeEvent(params []interface{}) map[string]interface{} {
	if !config.DefConfig.Common.EnableEventLog {
		return responsePack(berr.INVALID_METHOD, "")
	}
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}

	switch (params[0]).(type) {
	// block height
	case float64:
		height := uint32(params[0].(float64))
		eventInfos, err := bactor.GetEventNotifyByHeight(height)
		if err != nil {
			if err == scom.ErrNotFound {
				return responseSuccess(nil)
			}
			return responsePack(berr.INTERNAL_ERROR, "")
		}

		if len(params) == 2 {
			addr, ok := params[1].(string)
			if !ok {
				return responsePack(berr.INVALID_PARAMS, "")
			}
			address, err := bcomn.GetAddress(addr)
			if err != nil {
				return responsePack(berr.INVALID_PARAMS, "")
			}

			eInfos := []interface{}{}
			for _, eventInfo := range eventInfos {
				notify := bcomn.GetEventForContract(eventInfo, address)
				eInfos = append(eInfos, notify...)
			}
			return responseSuccess(eInfos)
		} else {
			eInfos := make([]*bcomn.ExecuteNotify, 0, len(eventInfos))
			for _, eventInfo := range eventInfos {
				_, notify := bcomn.GetExecuteNotify(eventInfo)
				eInfos = append(eInfos, &notify)
			}
			return responseSuccess(eInfos)
		}

		//txhash
	case string:
		str := params[0].(string)
		hash, err := common.Uint256FromHexString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		eventInfo, err := bactor.GetEventNotifyByTxHash(hash)
		if err != nil {
			if scom.ErrNotFound == err {
				return responseSuccess(nil)
			}
			return responsePack(berr.INTERNAL_ERROR, "")
		}
		_, notify := bcomn.GetExecuteNotify(eventInfo)
		return responseSuccess(notify)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responsePack(berr.INVALID_PARAMS, "")
}

//get block height by transaction hash
func GetBlockHeightByTxHash(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}

	switch (params[0]).(type) {
	// tx hash
	case string:
		str := params[0].(string)
		hash, err := common.Uint256FromHexString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		height, _, err := bactor.GetTxnWithHeightByTxHash(hash)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		return responseSuccess(height)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responsePack(berr.INVALID_PARAMS, "")
}

//get balance of address
func GetBalance(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	addrBase58, ok := params[0].(string)
	if !ok {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	address, err := common.AddressFromBase58(addrBase58)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	rsp, err := bcomn.GetBalance(address)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responseSuccess(rsp)
}

//get allowance
func GetAllowance(params []interface{}) map[string]interface{} {
	if len(params) < 3 {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	asset, ok := params[0].(string)
	if !ok {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	fromAddrStr, ok := params[1].(string)
	if !ok {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	fromAddr, err := bcomn.GetAddress(fromAddrStr)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	toAddrStr, ok := params[2].(string)
	if !ok {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	toAddr, err := bcomn.GetAddress(toAddrStr)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	rsp, err := bcomn.GetAllowance(asset, fromAddr, toAddr)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responseSuccess(rsp)
}

//get merkle proof by transaction hash
func GetMerkleProof(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	str, ok := params[0].(string)
	if !ok {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	hash, err := common.Uint256FromHexString(str)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	height, _, err := bactor.GetTxnWithHeightByTxHash(hash)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	header, err := bactor.GetHeaderByHeight(height)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}

	curHeight := bactor.GetCurrentBlockHeight()
	curHeader, err := bactor.GetHeaderByHeight(curHeight)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	proof, err := bactor.GetMerkleProof(uint32(height), uint32(curHeight))
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, "")
	}
	var hashes []string
	for _, v := range proof {
		hashes = append(hashes, v.ToHexString())
	}
	return responseSuccess(bcomn.MerkleProof{"MerkleProof", header.TransactionsRoot.ToHexString(), height,
		curHeader.BlockRoot.ToHexString(), curHeight, hashes})
}

//get block transactions by height
func GetBlockTxsByHeight(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	switch params[0].(type) {
	case float64:
		height := uint32(params[0].(float64))
		hash := bactor.GetBlockHashFromStore(height)
		if hash == common.UINT256_EMPTY {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		block, err := bactor.GetBlockFromStore(hash)
		if err != nil {
			return responsePack(berr.UNKNOWN_BLOCK, "")
		}
		return responseSuccess(bcomn.GetBlockTransactions(block))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
}

//get gas price in block
func GetGasPrice(params []interface{}) map[string]interface{} {
	result, err := bcomn.GetGasPrice()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, "")
	}
	return responseSuccess(result)
}

func CheckTorrent(params []interface{}) map[string]interface{} {
	var filehash string
	switch (params[0]).(type) {
	case string:
		filehash = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	if len(filehash) != consts.PROTO_NODE_FILE_HASH_LEN {
		return responsePack(berr.INVALID_PARAMS, fmt.Sprintf("invalid fileHash len is %d, not 46", len(filehash)))
	}

	peers, err := tracker.CheckTorrent(filehash)
	if err != nil {
		if err.Error() == "not found" {
			return responseSuccess(&httpComm.TorrentPeersRsp{Peers: peers})
		}
		return responsePack(berr.INTERNAL_ERROR, "")
	}
	return responseSuccess(&httpComm.TorrentPeersRsp{
		Peers: peers,
	})
	return nil
}

func EndPointReg(params []interface{}) map[string]interface{} {
	switch (params[0]).(type) {
	case string:
		wAddr = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	switch (params[1]).(type) {
	case string:
		host = params[1].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	if wAddr == "" || host == "" {
		return responsePack(berr.INVALID_PARAMS, "")
	}

	if err := tracker.EndPointRegistry(wAddr, host); err != nil {
		log.Errorf("EndPointRegistry error:%s\n", err)
		return responsePack(berr.INTERNAL_ERROR, "")
	}
	return responseSuccess(&httpComm.EndPointRsp{
		Wallet: wAddr,
		Host:   host,
	})
}

func EndPointUpdate(params []interface{}) map[string]interface{} {
	switch (params[0]).(type) {
	case string:
		wAddr = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	switch (params[1]).(type) {
	case string:
		host = params[1].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	if wAddr == "" || host == "" {
		return responsePack(berr.INVALID_PARAMS, "")
	}

	err := tracker.EndPointRegUpdate(wAddr, host)
	if err != nil && err.Error() == "not exist" {
		return responsePack(berr.ENDPOINT_NOT_FOUND, "")
	} else if err != nil {
		log.Errorf("EndPointUpdate error:%s\n", err)
		return responsePack(berr.INTERNAL_ERROR, "")
	}
	log.Debugf("rpc/interface/endpointupdate wAddr: %s, host:%s\n", wAddr, host)

	return responseSuccess(&httpComm.EndPointRsp{
		Wallet: wAddr,
		Host:   host,
	})
}

func EndPointUnReg(params []interface{}) map[string]interface{} {
	switch (params[0]).(type) {
	case string:
		wAddr = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	err := tracker.EndPointUnRegistry(wAddr)
	if err != nil && err.Error() == "not exist" {
		return responsePack(berr.ENDPOINT_NOT_FOUND, "")
	} else if err != nil {
		log.Errorf("EndPointUnRegistry error:%s\n", err)
		return responsePack(berr.INTERNAL_ERROR, "")
	}

	return responseSuccess(&httpComm.EndPointRsp{
		Wallet: wAddr,
	})
}

func EndPointReq(params []interface{}) map[string]interface{} {
	switch (params[0]).(type) {
	case string:
		wAddr = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	host, err := tracker.EndPointQuery(wAddr)
	if err != nil && err.Error() == "not found" {
		return responsePack(berr.ENDPOINT_NOT_FOUND, "")
	} else if err != nil {
		log.Errorf("EndPointReq error:%s\n", err)
		return responsePack(berr.INTERNAL_ERROR, "")
	}
	log.Debugf("rpc/interface/endpointreq host:%s\n", host)

	return responseSuccess(&httpComm.EndPointRsp{
		Wallet: wAddr,
		Host:   host,
	})
}

func RegisterDns(params []interface{}) map[string]interface{} {
	var ipstr string
	var portstr string
	var initDeposituint64 uint64
	switch (params[0]).(type) {
	case string:
		ipstr = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch (params[1]).(type) {
	case string:
		portstr = params[1].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch (params[2]).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		initDeposituint64 = uint64(params[2].(uint64))
	case float32, float64:
		// may be bugs
		initDeposituint64 = uint64(params[2].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	ret, err := service.ScanNode.DNSNodeReg(ipstr, portstr, initDeposituint64)
	if err != nil {
		log.Errorf("RegisterDns error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/registerdns ip:%s port:%s initDeposit:%d\n", ipstr, portstr, initDeposituint64)
	dspRsp := httpComm.DnsRsp{
		Tx: hex.EncodeToString(common.ToArrayReverse(ret.ToArray())),
	}
	return responseSuccess(&dspRsp)
}

func UnregisterDns(params []interface{}) map[string]interface{} {
	ret, err := service.ScanNode.DNSNodeUnreg()
	if err != nil {
		log.Errorf("UnRegisterDns error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/unregisterdns %s", "")
	dspRsp := httpComm.DnsRsp{
		Tx: hex.EncodeToString(common.ToArrayReverse(ret.ToArray())),
	}
	return responseSuccess(&dspRsp)
}

func QuitDns(params []interface{}) map[string]interface{} {
	ret, err := service.ScanNode.DNSNodeQuit()
	if err != nil {
		log.Errorf("QuitDns error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/quitdns %s", "")
	dspRsp := httpComm.DnsRsp{
		Tx: hex.EncodeToString(common.ToArrayReverse(ret.ToArray())),
	}
	return responseSuccess(&dspRsp)
}

func AddDnsPos(params []interface{}) map[string]interface{} {
	var deltaDeposit uint64
	switch (params[0]).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		deltaDeposit = uint64(params[0].(uint64))
	case float32, float64:
		// may be bugs
		deltaDeposit = uint64(params[0].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	ret, err := service.ScanNode.DNSAddPos(deltaDeposit)
	if err != nil {
		log.Errorf("AddDnsPos error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/adddnspos deltaDeposit:%d\n", deltaDeposit)
	dspRsp := httpComm.DnsRsp{
		Tx: hex.EncodeToString(common.ToArrayReverse(ret.ToArray())),
	}
	return responseSuccess(&dspRsp)
}

func ReduceDnsPos(params []interface{}) map[string]interface{} {
	var deltaDeposit uint64
	switch (params[0]).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		deltaDeposit = uint64(params[0].(uint64))
	case float32, float64:
		// may be bugs
		deltaDeposit = uint64(params[0].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	ret, err := service.ScanNode.DNSReducePos(deltaDeposit)
	if err != nil {
		log.Errorf("AddDnsPos error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/reducednspos deltaDeposit:%d\n", deltaDeposit)
	dspRsp := httpComm.DnsRsp{
		Tx: hex.EncodeToString(common.ToArrayReverse(ret.ToArray())),
	}
	return responseSuccess(&dspRsp)
}

func GetRegisterDnsInfo(params []interface{}) map[string]interface{} {
	var dnsAll bool
	var peerPubkey string
	switch (params[0]).(type) {
	case bool:
		dnsAll = params[0].(bool)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	switch (params[1]).(type) {
	case string:
		peerPubkey = params[1].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	if dnsAll {
		m, err := service.ScanNode.GetDnsPeerPoolMap()
		if err != nil {
			log.Errorf("Get all dns register info err:%s\n", err)
			return responsePack(berr.INTERNAL_ERROR, err.Error())
		}

		if _, ok := m.PeerPoolMap[""]; ok {
			delete(m.PeerPoolMap, "")
		}

		fmt.Println("rpc/interface/getregisterdnsinfo getdnspeeerpoolmap")
		dnsPPRsp := httpComm.DnsPeerPoolRsp{
			PeerPoolMap:  m.PeerPoolMap,
			PeerPoolItem: nil,
		}
		return responseSuccess(&dnsPPRsp)
	} else if peerPubkey != "" {
		item, err := service.ScanNode.GetDnsPeerPoolItem(peerPubkey)
		if err != nil {
			log.Errorf("Get all dns register info err:%s\n", err)
			return responsePack(berr.INTERNAL_ERROR, err.Error())
		}
		dnsPPRsp := httpComm.DnsPeerPoolRsp{
			PeerPoolMap:  nil,
			PeerPoolItem: item,
		}
		return responseSuccess(&dnsPPRsp)
	} else {
		return responsePack(berr.INVALID_PARAMS, "")
	}
}

func GetDnsHostInfo(params []interface{}) map[string]interface{} {
	var dnsAll bool
	var walletAddr string
	switch (params[0]).(type) {
	case bool:
		dnsAll = params[0].(bool)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	switch (params[1]).(type) {
	case string:
		walletAddr = params[1].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	if dnsAll {
		m, err := service.ScanNode.GetAllDnsNodes()
		if err != nil {
			log.Errorf("Get all dns register info err:%s\n", err)
			return responsePack(berr.INTERNAL_ERROR, err.Error())
		}

		if _, ok := m[""]; ok {
			delete(m, "")
		}

		fmt.Println("rpc/interface/getdnshostinfo GetAllDnsNodes")
		dnsNIRsp := httpComm.DnsNodeInfoRsp{
			NodeInfoMap:  m,
			NodeInfoItem: nil,
		}
		return responseSuccess(&dnsNIRsp)
	} else if walletAddr != "" {
		addr, err := common.AddressFromBase58(walletAddr)
		if err != nil {
			log.Errorf("Get dns host info err:%s\n", err)
			return responsePack(berr.INTERNAL_ERROR, err.Error())
		}

		item, err := service.ScanNode.GetDnsNodeByAddr(addr)
		if err != nil {
			log.Errorf("Get all dns register info err:%s\n", err)
			return responsePack(berr.INTERNAL_ERROR, err.Error())
		}

		fmt.Println("rpc/interface/getdnshostinfo GetDnsNodeByAddr")
		dnsNIRsp := httpComm.DnsNodeInfoRsp{
			NodeInfoMap:  nil,
			NodeInfoItem: item,
		}
		return responseSuccess(&dnsNIRsp)
	} else {
		return responsePack(berr.INVALID_PARAMS, "")
	}
}

func OpenChannel(params []interface{}) map[string]interface{} {
	var partnerAddrstr string
	switch (params[0]).(type) {
	case string:
		partnerAddrstr = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	host, err := service.ScanNode.QueryHostInfo(partnerAddrstr)
	if host == "" {
		log.Errorf("OpenChannel hostInfo is null, error: %s", err)
		return responsePack(berr.CHANNEL_TARGET_HOST_INFO_NOT_FOUND, "")
	}

	id, err := service.ScanNode.OpenChannel(partnerAddrstr, 0)
	if err != nil {
		log.Errorf("OpenChannel error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/openchannel partneraddr:%s\n", partnerAddrstr)
	channelRsp := httpComm.ChannelRsp{
		Id: uint32(id),
	}
	return responseSuccess(&channelRsp)
}

func CloseChannel(params []interface{}) map[string]interface{} {
	var partnerAddrstr string
	switch (params[0]).(type) {
	case string:
		partnerAddrstr = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	host, err := service.ScanNode.QueryHostInfo(partnerAddrstr)
	if host == "" {
		log.Errorf("CloseChannel hostInfo is null, error: %s", err)
		return responsePack(berr.CHANNEL_TARGET_HOST_INFO_NOT_FOUND, "")
	}

	err = service.ScanNode.CloseChannel(partnerAddrstr)
	if err != nil {
		log.Errorf("CloseChannel error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/closechannel partneraddr:%s\n", partnerAddrstr)

	return responseSuccess(&httpComm.SuccessRsp{})
}

func DepositToChannel(params []interface{}) map[string]interface{} {
	var partnerAddress string
	var deposit uint64
	switch (params[0]).(type) {
	case string:
		partnerAddress = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch (params[1]).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		deposit = params[1].(uint64)
	case float32, float64:
		deposit = uint64(params[1].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	host, err := service.ScanNode.QueryHostInfo(partnerAddress)
	if host == "" {
		log.Errorf("DepositToChannel hostInfo is null, error: %s", err)
		return responsePack(berr.CHANNEL_TARGET_HOST_INFO_NOT_FOUND, "")
	}

	balance, err := service.ScanNode.QueryChannelDeposit(partnerAddress)
	if err != nil {
		log.Errorf("DepositToChannel error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}

	deposit += balance
	err = service.ScanNode.DepositToChannel(partnerAddress, deposit)
	if err != nil {
		log.Errorf("DepositToChannel error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/depositchannel partneraddr:%s totaldeposit:%d \n", partnerAddress, deposit)
	return responseSuccess(&httpComm.SuccessRsp{})
}

func TransferToSomebody(params []interface{}) map[string]interface{} {
	var partnerAddrstr string
	var amountuint64 uint64
	var paymentIduint int32
	switch (params[0]).(type) {
	case string:
		partnerAddrstr = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch (params[1]).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		amountuint64 = params[1].(uint64)
	case float32, float64:
		// may be bugs
		amountuint64 = uint64(params[1].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch (params[2]).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		paymentIduint = int32(params[2].(uint64))
	case float32, float64:
		// may be bugs
		paymentIduint = int32(params[2].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	host, err := service.ScanNode.QueryHostInfo(partnerAddrstr)
	if host == "" {
		log.Errorf("TransferToSomebody hostinfo is null, error: %s", err)
		return responsePack(berr.CHANNEL_TARGET_HOST_INFO_NOT_FOUND, "")
	}

	err = service.ScanNode.Transfer(paymentIduint, amountuint64, partnerAddrstr)
	if err != nil {
		log.Errorf("TransferToSomebody error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/transferchannel partneraddr:%s amount:%d paymentid:%d\n", partnerAddrstr, amountuint64, paymentIduint)
	return responseSuccess(&httpComm.SuccessRsp{})
}

func MediaTransferToSomebody(params []interface{}) map[string]interface{} {
	var paymentIduint int32
	var amountuint64 uint64
	var mediaAddrStr string
	var partnerAddrstr string
	switch (params[0]).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		paymentIduint = int32(params[0].(uint64))
	case float32, float64:
		// may be bugs
		paymentIduint = int32(params[0].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch (params[1]).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		amountuint64 = params[1].(uint64)
	case float32, float64:
		// may be bugs
		amountuint64 = uint64(params[1].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch (params[2]).(type) {
	case string:
		mediaAddrStr = params[2].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch (params[3]).(type) {
	case string:
		partnerAddrstr = params[3].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	host, err := service.ScanNode.QueryHostInfo(partnerAddrstr)
	if host == "" {
		log.Errorf("MediaTransferToSomebody hostinfo is null, error: %s", err)
		return responsePack(berr.CHANNEL_TARGET_HOST_INFO_NOT_FOUND, "")
	}

	err = service.ScanNode.MediaTransfer(paymentIduint, amountuint64, mediaAddrStr, partnerAddrstr)
	if err != nil {
		log.Errorf("MediaTransferToSomebody error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/mediatransferchannel mediaaddr:%s partneraddr:%s amount:%d paymentid:%d\n", mediaAddrStr, partnerAddrstr, amountuint64, paymentIduint)
	return responseSuccess(&httpComm.SuccessRsp{})
}

func WithdrawChannel(params []interface{}) map[string]interface{} {
	var partnerAddr string
	var withdraw uint64
	switch (params[0]).(type) {
	case string:
		partnerAddr = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch (params[1]).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		withdraw = params[1].(uint64)
	case float32, float64:
		withdraw = uint64(params[1].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	host, err := service.ScanNode.QueryHostInfo(partnerAddr)
	if host == "" {
		log.Errorf("WithdrawChannel hostInfo is null, error: %s", err)
		return responsePack(berr.CHANNEL_TARGET_HOST_INFO_NOT_FOUND, "")
	}

	balance, err := service.ScanNode.QueryChannelWithdraw(partnerAddr)
	if err != nil {
		log.Errorf("DepositToChannel error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}

	withdraw += balance
	err = service.ScanNode.WithdrawFromChannel(partnerAddr, withdraw)
	if err != nil {
		log.Errorf("WithdrawChannel error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/withdrawchannel partneraddr:%s amount:%d \n", partnerAddr, withdraw)
	return responseSuccess(&httpComm.SuccessRsp{})
}

func GetAllChannels(params []interface{}) map[string]interface{} {
	channelInfos, err := service.ScanNode.GetAllChannels()
	if err != nil {
		log.Errorf("GetAllChannels error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	return responseSuccess(channelInfos)
}

func QueryChannelDeposit(params []interface{}) map[string]interface{} {
	var partnerAddrstr string
	switch (params[0]).(type) {
	case string:
		partnerAddrstr = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	balance, err := service.ScanNode.QueryChannelDeposit(partnerAddrstr)
	if err != nil {
		log.Errorf("QueryChannelDeposit error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/openchanneldeposit partneraddr:%s\n", partnerAddrstr)
	curBalanceRsp := httpComm.ChannelTotalDepositBalanceRsp{
		TotalDepositBalance: balance,
		TotalDepositBalanceFormat: utils.FormatUsdt(balance),
	}
	return responseSuccess(&curBalanceRsp)
}

func QueryHostInfo(params []interface{}) map[string]interface{} {
	var partnerAddrstr string
	switch (params[0]).(type) {
	case string:
		partnerAddrstr = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	host, err := service.ScanNode.QueryHostInfo(partnerAddrstr)
	if err != nil {
		log.Errorf("QueryHostInfo error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/queryhostinfo partneraddr:%s\n", partnerAddrstr)
	endpointRsp := httpComm.EndPointRsp{
		Wallet: partnerAddrstr,
		Host:   host,
	}
	return responseSuccess(&endpointRsp)
}

func GetChannelInitProgress(params []interface{}) map[string]interface{} {
	progress, err := service.ScanNode.GetFilterBlockProgress()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	progressRsp := httpComm.FilterBlockProgress{
		Progress: progress.Progress,
		Start:    progress.Start,
		End:      progress.End,
		Now:      progress.Now,
	}
	return responseSuccess(&progressRsp)
}

func JoinDnsNodesChannels(params []interface{}) map[string]interface{} {
	progress, err := service.ScanNode.GetFilterBlockProgress()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	if progress.Progress != 1.0 {
		return responsePack(berr.BLOCK_SYNCING_UNCOMPLETE, errors.New("block sync uncomplete, please wait a moment."))
	}

	err = service.ScanNode.AutoSetupDNSChannelsWorking()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}

	return responseSuccess(nil)
}

func CooperativeSettle(params []interface{}) map[string]interface{} {
	var partnerAddress string
	switch (params[0]).(type) {
	case string:
		partnerAddress = params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	err := service.ScanNode.CooperativeSettle(partnerAddress)
	if err != nil {
		log.Errorf("CooperativeSettle error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/CooperativeSettle\n")
	return nil
}

func GetFee(params []interface{}) map[string]interface{} {
	var channelID uint64

	switch (params[0]).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		channelID = params[0].(uint64)
	case float32, float64:
		channelID = uint64(params[0].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	fee, err := service.ScanNode.GetFee(channelID)
	if err != nil {
		log.Errorf("GetFee error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/getfee\n")
	curBalanceRsp := httpComm.ChannelFeeRsp{
		Flat:               uint64(fee.Flat),
		Proportional:       uint64(fee.Proportional),
		FlatFormat:         utils.FormatUsdt(uint64(fee.Flat)),
		ProportionalFormat: utils.FormatUsdt(uint64(fee.Proportional)),
	}
	return responseSuccess(&curBalanceRsp)
}

func SetFee(params []interface{}) map[string]interface{} {
	var flat uint64
	var pro uint64

	switch (params[0]).(type) {
	case float64:
		flat = uint64(params[0].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch (params[1]).(type) {
	case float64:
		pro = uint64(params[1].(float64))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	fee := &transfer.FeeScheduleState{
		Flat: chanCom.FeeAmount(flat),
		Proportional: chanCom.ProportionalFeeAmount(pro),
	}
	err := service.ScanNode.SetFee(fee)
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, "")
	}

	return nil
}

func GetPenalty(params []interface{}) map[string]interface{} {
	res, err := service.ScanNode.GetPenalty()
	if err != nil {
		log.Errorf("GetPenalty error: %s", err)
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	fmt.Printf("rpc/interface/GetPenalty\n")
	return responseSuccess(&res)
}

func SetPenalty(params []interface{}) map[string]interface{} {
	var feePenalty float64
	var diversityPenalty float64

	switch (params[0]).(type) {
	case float64:
		feePenalty = params[0].(float64)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch (params[1]).(type) {
	case float64:
		diversityPenalty = params[1].(float64)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	penalty := &chanCom.RoutePenaltyConfig{
		FeePenalty:       feePenalty,
		DiversityPenalty: diversityPenalty,
	}
	err := service.ScanNode.SetPenalty(penalty)
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, "")
	}

	return nil
}