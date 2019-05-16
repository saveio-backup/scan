/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-15
 */
package utils

import (
	"encoding/json"
	"fmt"

	"github.com/saveio/dsp-go-sdk/channel"
	httpComm "github.com/saveio/scan/http/base/common"
	berr "github.com/saveio/scan/http/base/error"
	"github.com/saveio/themis/common/log"
)

func RegEndPoint(waddr, host string) (*httpComm.EndPointRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("regendpoint", []interface{}{waddr, host})
	endPoint := &httpComm.EndPointRsp{}
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid address: %s", waddr),
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

	err := json.Unmarshal(result, endPoint)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("DDNS endpoint register success wallet:%s,host%s", endPoint.Wallet, endPoint.Host)
	log.Debugf("RegEndPoint result :%s", result)
	return endPoint, nil
}

func UpdateEndPoint(waddr, host string) (*httpComm.EndPointRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("updateendpoint", []interface{}{waddr, host})
	endPoint := &httpComm.EndPointRsp{}
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid address: %s, host: %s", waddr, host),
			}
		case berr.ENDPOINT_NOT_FOUND:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.ENDPOINT_NOT_FOUND,
				ErrMsg:    berr.ErrMap[berr.ENDPOINT_NOT_FOUND],
				FailedMsg: fmt.Sprintf("This wallet address %s did not register endpoint, please reg first.", waddr),
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
	log.Debugf("DDNS endpoint update success wallet:%s,host%s", endPoint.Wallet, endPoint.Host)
	log.Debugf("UpdateEndPoint result :%s", result)
	return endPoint, nil
}

func UnRegEndPoint(waddr string) (*httpComm.EndPointRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("unRegendpoint", []interface{}{waddr})
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
				FailedMsg: fmt.Sprintf("This wallet address %s did not register endpoint, can not unregister.", waddr),
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
	log.Debugf("DDNS endpoint update unregister wallet:%s", endPoint.Wallet)
	log.Debugf("UnRegEndPoint result :%s", result)
	return endPoint, nil
}

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

func OpenChannel(partnerAddr string) (*httpComm.ChannelRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("openchannel", []interface{}{partnerAddr})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid partnerAddr: %s", partnerAddr),
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

	chanRsp := &httpComm.ChannelRsp{}
	err := json.Unmarshal(result, chanRsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("openchannel success")
	log.Debugf("OpenChannel result :%s", result)
	return chanRsp, nil
}

func DepositToChannel(partnerAddr string, totalDeposit uint64) (*httpComm.SuccessRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("depositchannel", []interface{}{partnerAddr, totalDeposit})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid partnerAddr: %s, totalDeposit: %d", partnerAddr, totalDeposit),
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

	log.Debugf("depositchannel success")
	log.Debugf("DepositToChannel result :%s", result)
	return nil, nil
}

func TransferToSomebody(partnerAddr string, amount uint64, paymentid uint) (*httpComm.SuccessRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("transferchannel", []interface{}{partnerAddr, amount, paymentid})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid partnerAddr: %s, amount: %d, paymentid: %d", partnerAddr, amount, paymentid),
			}
		case berr.CHANNEL_TARGET_HOST_INFO_NOT_FOUND:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.CHANNEL_TARGET_HOST_INFO_NOT_FOUND,
				ErrMsg:    berr.ErrMap[berr.CHANNEL_TARGET_HOST_INFO_NOT_FOUND],
				FailedMsg: fmt.Sprintf("Address: %s hostinfo not found", partnerAddr),
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

	log.Debugf("transferchannel success")
	log.Debugf("TransferToSomebody result :%s", result)
	return nil, nil
}

func QueryChannelDeposit(partnerAddr string) (*httpComm.ChannelTotalDepositBalanceRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("querychanneldeposit", []interface{}{partnerAddr})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid partnerAddr: %s", partnerAddr),
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
	totalDepositBalanceRsp := &httpComm.ChannelTotalDepositBalanceRsp{}
	err := json.Unmarshal(result, totalDepositBalanceRsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("querychanneldeposit success")
	log.Debugf("QueryChannelDeposit result :%s", result)
	return totalDepositBalanceRsp, nil
}

func WithdrawChannel(partnerAddr string, totalDeposit uint64) (*httpComm.SuccessRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("withdrawchannel", []interface{}{partnerAddr, totalDeposit})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid partnerAddr: %s, totalDeposit: %d", partnerAddr, totalDeposit),
			}
		case berr.CHANNEL_TARGET_HOST_INFO_NOT_FOUND:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.CHANNEL_TARGET_HOST_INFO_NOT_FOUND,
				ErrMsg:    berr.ErrMap[berr.CHANNEL_TARGET_HOST_INFO_NOT_FOUND],
				FailedMsg: fmt.Sprintf("Target address: %s host info not found", partnerAddr),
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

	log.Debugf("withdrawchannel success")
	log.Debugf("WithdrawChannel result :%s", result)
	return nil, nil
}

func GetAllChannels() (*channel.ChannelInfosResp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("getallchannels", []interface{}{})
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

	channelInfos := &channel.ChannelInfosResp{}
	err := json.Unmarshal(result, channelInfos)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}

	log.Debugf("getallchannels success")
	log.Debugf("getallchannels result :%s", result)
	return channelInfos, nil
}

func GetCurrentBalance(partnerAddr string) (*httpComm.ChannelCurrentBalanceRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("getcurrentbalance", []interface{}{partnerAddr})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid partnerAddr: %s", partnerAddr),
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
	currBalanceRsp := &httpComm.ChannelCurrentBalanceRsp{}
	err := json.Unmarshal(result, currBalanceRsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("GetCurrentBalance success")
	log.Debugf("GetCurrentBalance result :%s", result)
	return currBalanceRsp, nil
}

func QueryHostInfo(partnerAddr string) (*httpComm.EndPointRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("queryhostinfo", []interface{}{partnerAddr})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid partnerAddr: %s", partnerAddr),
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
	chanHostRsp := &httpComm.EndPointRsp{}
	err := json.Unmarshal(result, chanHostRsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("QueryChannelHostInfo success")
	log.Debugf("QueryChannelHostInfo result :%s", result)
	return chanHostRsp, nil
}
