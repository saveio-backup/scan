/*
 * Copyright (C) 2019 The oniChain Authors
 * This file is part of The oniChain library.
 *
 * The oniChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The oniChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The oniChain.  If not, see <http://www.gnu.org/licenses/>.
 */

package rpc

import (
	"os"
	"path/filepath"

	"github.com/oniio/oniChain/common/log"
	berr "github.com/oniio/oniDNS/http/base/error"
	"github.com/oniio/oniDNS/tracker"
	"github.com/oniio/oniChain/account"
	httpComm "github.com/oniio/oniDNS/http/base/common"
	"encoding/json"
	"github.com/oniio/oniChain/core/signature"
	"fmt"
)

const (
	RANDBYTELEN = 4
)

var (
	wAddr,host string
	sigData []byte
	acc *account.Account
	
)
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}




func SetDebugInfo(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	switch params[0].(type) {
	case float64:
		level := params[0].(float64)
		if err := log.Log.SetDebugLevel(int(level)); err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responsePack(berr.SUCCESS, true)
}

func EndPointReg(params []interface{}) map[string]interface{}{
	log.Debugf("in endreg:%v\n",params)
	fmt.Println("in endreg")
	switch (params[0]).(type) {
	case string:
		wAddr=params[0].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	switch (params[1]).(type) {
	case string:
		host=params[1].(string)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	log.Debugf("EndPointReg wallet:%s,host:%s",wAddr,host)
	endPoint:=httpComm.EndPointRsp{
		Wallet:wAddr,
		Host:host,
	}
	em,err:=json.Marshal(endPoint)
	if err!=nil{
		return responsePack(berr.JSON_MARSHAL_ERROR,"")
	}
	switch (params[2]).(type) {
	case []byte:
		sigData=params[2].([]byte)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	switch (params[3]).(type) {
	case *account.Account:
		acc=params[3].(*account.Account)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	if err:= signature.Verify(acc.PublicKey,em,sigData);err!=nil{
		return responsePack(berr.SIG_VERIFY_ERROR,"")
	}
	if err=tracker.EndPointRegistry(wAddr,host);err!=nil{
		//PrintErrorMsg("regEndPoint EndPointRegistry error:%s\n",err)
		log.Errorf("regEndPoint EndPointRegistry error:%s\n",err)
		return responsePack(berr.INTERNAL_ERROR,"")
	}
	//return responsePack(berr.SUCCESS,true)
	return responseSuccess(&endPoint)
}