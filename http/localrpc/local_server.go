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

// Package localrpc privides a function to start local rpc server
package localrpc

import (
	"net/http"
	"strconv"

	"fmt"
	"github.com/saveio/themis/common/log"
	cfg "github.com/saveio/scan/common/config"
	"github.com/saveio/scan/http/base/rpc"
)

const (
	LOCAL_HOST string = "127.0.0.1"
	//LOCAL_DIR  string = "/local"
	LOCAL_DIR string = "/"
)

func StartLocalServer() error {
	log.Info("start local Server")
	http.HandleFunc(LOCAL_DIR, rpc.Handle)
	rpc.HandleFunc("setdebuginfo", rpc.SetDebugInfo)
	//rpc.HandleFunc("regendpoint",rpc.EndPointReg)

	// TODO: only listen to local host
	err := http.ListenAndServe(":"+strconv.Itoa(int(cfg.DefaultConfig.RpcConfig.HttpLocalPort)), nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	return nil
}
