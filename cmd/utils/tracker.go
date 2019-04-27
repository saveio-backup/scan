/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-15
 */
package utils

import (
	"encoding/json"
	"fmt"
	"github.com/oniio/oniChain/common/log"
	httpComm "github.com/oniio/oniDNS/http/base/common"
)

func RegEndPoint(waddr, host string) error {
	result, ontErr := sendRpcRequest("regendpoint", []interface{}{waddr, host})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return fmt.Errorf("invalid address:%s,host:%s", waddr, host)
		}
		return ontErr.Error
	}
	endPoint := &httpComm.EndPointRsp{}

	err := json.Unmarshal(result, endPoint)
	if err != nil {
		return fmt.Errorf("json.Unmarshal error:%s", err)
	}
	log.Infof("DDNS endpoint registed success wallet:%s,host%s", endPoint.Wallet, endPoint.Host)
	log.Infof("RegEndPoint result :%s", result)
	return nil
}

func UpdateEndPoint(waddr, host string) error {
	result, ontErr := sendRpcRequest("updateendpoint", []interface{}{waddr, host})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return fmt.Errorf("invalid address:%s,host:%s", waddr, host)
		}
		return ontErr.Error
	}
	endPoint := &httpComm.EndPointRsp{}

	err := json.Unmarshal(result, endPoint)
	if err != nil {
		return fmt.Errorf("json.Unmarshal error:%s", err)
	}
	log.Infof("DDNS endpoint update success wallet:%s,host%s", endPoint.Wallet, endPoint.Host)
	log.Infof("Update EndPoint result :%s", result)
	return nil
}

func UnRegEndPoint(waddr string) error {
	result, ontErr := sendRpcRequest("unRegendpoint", []interface{}{waddr})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return fmt.Errorf("invalid address:%s", waddr)
		}
		return ontErr.Error
	}
	endPoint := &httpComm.EndPointRsp{}

	err := json.Unmarshal(result, endPoint)
	if err != nil {
		return fmt.Errorf("json.Unmarshal error:%s", err)
	}
	log.Infof("DDNS endpoint unRegisted success wallet:%s,host%s", endPoint.Wallet)
	log.Infof("unRegEndPoint result :%s", result)
	return nil
}

func ReqEndPoint(waddr string) error {
	result, ontErr := sendRpcRequest("reqendpoint", []interface{}{waddr})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return fmt.Errorf("invalid address:%s", waddr)
		}
		return ontErr.Error
	}
	endPoint := &httpComm.EndPointRsp{}

	err := json.Unmarshal(result, endPoint)
	if err != nil {
		return fmt.Errorf("json.Unmarshal error:%s", err)
	}
	log.Infof("DDNS endpoint request success wallet:%s,host%s", endPoint.Wallet, endPoint.Host)
	log.Infof("ReqEndPoint result :%s", result)
	return nil
}
