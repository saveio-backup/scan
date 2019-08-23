package utils

import (
	"encoding/json"
	"fmt"

	httpComm "github.com/saveio/scan/http/base/common"
	berr "github.com/saveio/scan/http/base/error"
	"github.com/saveio/themis/common/log"
)

func DNSNodeReg(ip string, port string, initDeposit uint64) (*httpComm.DnsRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("registerdns", []interface{}{ip, port, initDeposit})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid ip: %s, port: %s, initDeposit: %d", ip, port, initDeposit),
			}
		}
		return nil, &httpComm.FailedRsp{
			ErrCode:   ontErr.ErrorCode,
			ErrMsg:    "",
			FailedMsg: ontErr.Error.Error(),
		}
	}

	dnsRsp := &httpComm.DnsRsp{}

	err := json.Unmarshal(result, dnsRsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("registerdns success")
	log.Debugf("DNSNodeReg result :%s", result)
	return dnsRsp, nil
}

func DNSNodeUnreg() (*httpComm.DnsRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("unregisterdns", []interface{}{})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case berr.INTERNAL_ERROR:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INTERNAL_ERROR,
				ErrMsg:    berr.ErrMap[berr.INTERNAL_ERROR],
				FailedMsg: ontErr.Error.Error(),
			}
		}
		return nil, &httpComm.FailedRsp{
			ErrCode:   ontErr.ErrorCode,
			ErrMsg:    "",
			FailedMsg: ontErr.Error.Error(),
		}
	}

	dnsRsp := &httpComm.DnsRsp{}

	err := json.Unmarshal(result, dnsRsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("unregisterdns success")
	log.Debugf("DNSNodeUnreg result :%s", result)
	return dnsRsp, nil
}

func DNSNodeQuit() (*httpComm.DnsRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("quitdns", []interface{}{})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case berr.INTERNAL_ERROR:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INTERNAL_ERROR,
				ErrMsg:    berr.ErrMap[berr.INTERNAL_ERROR],
				FailedMsg: ontErr.Error.Error(),
			}
		}
		return nil, &httpComm.FailedRsp{
			ErrCode:   ontErr.ErrorCode,
			ErrMsg:    "",
			FailedMsg: ontErr.Error.Error(),
		}
	}

	dnsRsp := &httpComm.DnsRsp{}

	err := json.Unmarshal(result, dnsRsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("quitdns success")
	log.Debugf("DNSNodeQuit result :%s", result)
	return dnsRsp, nil
}

func DNSAddPos(deltaDeposit uint64) (*httpComm.DnsRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("adddnspos", []interface{}{deltaDeposit})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid deltaDeposit: %d", deltaDeposit),
			}
		case berr.INTERNAL_ERROR:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INTERNAL_ERROR,
				ErrMsg:    berr.ErrMap[berr.INTERNAL_ERROR],
				FailedMsg: ontErr.Error.Error(),
			}
		}
		return nil, &httpComm.FailedRsp{
			ErrCode:   ontErr.ErrorCode,
			ErrMsg:    "",
			FailedMsg: ontErr.Error.Error(),
		}
	}
	dnsRsp := &httpComm.DnsRsp{}

	err := json.Unmarshal(result, dnsRsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("adddnspos success")
	log.Debugf("DNSAddPos result :%s", result)
	return dnsRsp, nil
}

func DNSReducePos(deltaDeposit uint64) (*httpComm.DnsRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("reducednspos", []interface{}{deltaDeposit})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid deltaDeposit: %d", deltaDeposit),
			}
		case berr.INTERNAL_ERROR:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INTERNAL_ERROR,
				ErrMsg:    berr.ErrMap[berr.INTERNAL_ERROR],
				FailedMsg: ontErr.Error.Error(),
			}
		}
		return nil, &httpComm.FailedRsp{
			ErrCode:   ontErr.ErrorCode,
			ErrMsg:    "",
			FailedMsg: ontErr.Error.Error(),
		}
	}
	dnsRsp := &httpComm.DnsRsp{}

	err := json.Unmarshal(result, dnsRsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("reducednspos success")
	log.Debugf("DNSReducePos result :%s", result)
	return dnsRsp, nil
}

func DNSRegisterInfo(dnsAll bool, peerPubKey string) (*httpComm.DnsPeerPoolRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("getregisterdnsinfo", []interface{}{dnsAll, peerPubKey})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid dnsAll: %t, peerPubKey: %s", dnsAll, peerPubKey),
			}
		case berr.INTERNAL_ERROR:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INTERNAL_ERROR,
				ErrMsg:    berr.ErrMap[berr.INTERNAL_ERROR],
				FailedMsg: ontErr.Error.Error(),
			}
		}
		return nil, &httpComm.FailedRsp{
			ErrCode:   ontErr.ErrorCode,
			ErrMsg:    "",
			FailedMsg: ontErr.Error.Error(),
		}
	}
	dnsPPRsp := &httpComm.DnsPeerPoolRsp{}

	err := json.Unmarshal(result, dnsPPRsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("getregisterdnsinfo success")
	log.Debugf("DNSRegisterInfo result :%s", result)
	return dnsPPRsp, nil
}

func DNSHostInfo(dnsAll bool, walletAddr string) (*httpComm.DnsNodeInfoRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("getdnshostinfo", []interface{}{dnsAll, walletAddr})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid dnsAll: %t, walletAddr: %s", dnsAll, walletAddr),
			}
		case berr.INTERNAL_ERROR:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INTERNAL_ERROR,
				ErrMsg:    berr.ErrMap[berr.INTERNAL_ERROR],
				FailedMsg: ontErr.Error.Error(),
			}
		}
		return nil, &httpComm.FailedRsp{
			ErrCode:   ontErr.ErrorCode,
			ErrMsg:    "",
			FailedMsg: ontErr.Error.Error(),
		}
	}
	dnsNIRsp := &httpComm.DnsNodeInfoRsp{}

	err := json.Unmarshal(result, dnsNIRsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("getdnshostinfo success")
	log.Debugf("DNSHostInfo result :%s", result)
	return dnsNIRsp, nil
}

func ReqEndPoint(waddr string) (*httpComm.EndPointRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("reqendpoint", []interface{}{waddr})
	endPoint := &httpComm.EndPointRsp{}
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid address: %s", waddr),
			}
		case berr.ENDPOINT_NOT_FOUND:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.ENDPOINT_NOT_FOUND,
				ErrMsg:    berr.ErrMap[berr.ENDPOINT_NOT_FOUND],
				FailedMsg: fmt.Sprintf("This wallet address %s did not register endpoint", waddr),
			}
		}
		return nil, &httpComm.FailedRsp{
			ErrCode:   ontErr.ErrorCode,
			ErrMsg:    "",
			FailedMsg: ontErr.Error.Error(),
		}
	}

	err := json.Unmarshal(result, endPoint)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("DDNS endpoint request success wallet:%s,host%s", endPoint.Wallet, endPoint.Host)
	log.Debugf("ReqEndPoint result :%s", result)
	return endPoint, nil
}
