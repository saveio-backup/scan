package utils

import (
	"encoding/json"
	"fmt"

	ch_actor "github.com/saveio/pylons/actor/server"
	httpComm "github.com/saveio/scan/http/base/common"
	berr "github.com/saveio/scan/http/base/error"
	"github.com/saveio/themis/common/log"
)

func CheckChannelInitProgress() (*httpComm.FilterBlockProgress, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("initprogress", []interface{}{})
	if ontErr != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   ontErr.ErrorCode,
			ErrMsg:    "",
			FailedMsg: ontErr.Error.Error(),
		}
	}

	progress := &httpComm.FilterBlockProgress{}
	err := json.Unmarshal(result, progress)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("openchannel success")
	log.Debugf("OpenChannel result :%s", result)
	return progress, nil
}

func JoinDnsChannels() *httpComm.FailedRsp {
	result, ontErr := sendRpcRequest("joindnschannels", []interface{}{})
	if ontErr != nil {
		return &httpComm.FailedRsp{
			ErrCode:   ontErr.ErrorCode,
			ErrMsg:    "",
			FailedMsg: ontErr.Error.Error(),
		}
	}

	log.Debugf("JoinDnsChannels result :%s", result)
	return nil
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

func CloseChannel(partnerAddr string) (*httpComm.SuccessRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("closechannel", []interface{}{partnerAddr})
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

	log.Debugf("closechannel success")
	log.Debugf("CloseChannel result :%s", result)
	return nil, nil
}

func DepositToChannel(partnerAddr string, totalDeposit uint64) (*httpComm.SuccessRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("depositchannel", []interface{}{partnerAddr, totalDeposit})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid params: %s, totalDeposit: %d", partnerAddr, totalDeposit),
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

func MediaTransferToSomebody(paymentId uint, amount uint64, mediaAddr string, partnerAddr string) (*httpComm.SuccessRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("mediatransferchannel", []interface{}{paymentId, amount, mediaAddr, partnerAddr})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid partnerAddr: %s, amount: %d, paymentid: %d", partnerAddr, amount, paymentId),
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
	log.Debugf("mediatransferchannel success")
	log.Debugf("MediaTransferToSomebody result :%s", result)
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
				FailedMsg: fmt.Sprintf("Invalid params: %s", ontErr),
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

func GetAllChannels() (*ch_actor.ChannelsInfoResp, *httpComm.FailedRsp) {
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

	channelInfos := &ch_actor.ChannelsInfoResp{}
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

func CooperativeSettle(partnerAddress string) *httpComm.FailedRsp {
	result, ontErr := sendRpcRequest("cooperativeSettle", []interface{}{partnerAddress})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid params"),
			}
		case berr.INTERNAL_ERROR:
			return &httpComm.FailedRsp{
				ErrCode:   berr.INTERNAL_ERROR,
				ErrMsg:    berr.ErrMap[berr.INTERNAL_ERROR],
				FailedMsg: ontErr.Error.Error(),
			}
		}
		return &httpComm.FailedRsp{
			ErrCode:   ontErr.ErrorCode,
			ErrMsg:    "",
			FailedMsg: ontErr.Error.Error(),
		}
	}
	rsp := &httpComm.ChannelFeeRsp{}
	err := json.Unmarshal(result, rsp)
	if err != nil {
		return &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("CooperativeSettle fee success")
	log.Debugf("CooperativeSettle fee result :%s", result)
	return nil
}

func GetFee(channelID uint64) (*httpComm.ChannelFeeRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("getfee", []interface{}{channelID})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid channelId"),
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
	rsp := &httpComm.ChannelFeeRsp{}
	err := json.Unmarshal(result, rsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("Get fee success")
	log.Debugf("Get fee result :%s", result)
	return rsp, nil
}

func SetFee(flat uint64, pro uint64) *httpComm.FailedRsp {
	result, err := sendRpcRequest("setfee", []interface{}{flat, pro})
	if err != nil {
		switch err.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid params: %d", flat),
			}
		case berr.INTERNAL_ERROR:
			return &httpComm.FailedRsp{
				ErrCode:   berr.INTERNAL_ERROR,
				ErrMsg:    berr.ErrMap[berr.INTERNAL_ERROR],
				FailedMsg: err.Error.Error(),
			}
		}
		return &httpComm.FailedRsp{
			ErrCode:   err.ErrorCode,
			ErrMsg:    "",
			FailedMsg: err.Error.Error(),
		}
	}
	log.Debugf("Set fee success")
	log.Debugf("Set fee result :%s", result)
	return nil
}

func GetPenalty() (*httpComm.ChannelPenaltyRsp, *httpComm.FailedRsp) {
	result, ontErr := sendRpcRequest("getPenalty", []interface{}{})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid channelId"),
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
	rsp := &httpComm.ChannelPenaltyRsp{}
	err := json.Unmarshal(result, rsp)
	if err != nil {
		return nil, &httpComm.FailedRsp{
			ErrCode:   berr.JSON_UNMARSHAL_ERROR,
			ErrMsg:    berr.ErrMap[berr.JSON_UNMARSHAL_ERROR],
			FailedMsg: err.Error(),
		}
	}
	log.Debugf("Get penalty success")
	log.Debugf("Get penalty result :%s", result)
	return rsp, nil
}

func SetPenalty(fp float64, dp float64) *httpComm.FailedRsp {
	result, err := sendRpcRequest("setPenalty", []interface{}{fp, dp})
	if err != nil {
		switch err.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return &httpComm.FailedRsp{
				ErrCode:   berr.INVALID_PARAMS,
				ErrMsg:    berr.ErrMap[berr.INVALID_PARAMS],
				FailedMsg: fmt.Sprintf("Invalid params"),
			}
		case berr.INTERNAL_ERROR:
			return &httpComm.FailedRsp{
				ErrCode:   berr.INTERNAL_ERROR,
				ErrMsg:    berr.ErrMap[berr.INTERNAL_ERROR],
				FailedMsg: err.Error.Error(),
			}
		}
		return &httpComm.FailedRsp{
			ErrCode:   err.ErrorCode,
			ErrMsg:    "",
			FailedMsg: err.Error.Error(),
		}
	}
	log.Debugf("Set penalty success")
	log.Debugf("Set penalty result :%s", result)
	return nil
}
