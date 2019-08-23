package utils

import (
	"encoding/json"
	"fmt"

	httpComm "github.com/saveio/scan/http/base/common"
	berr "github.com/saveio/scan/http/base/error"
	"github.com/saveio/themis/common/log"
)

func CheckTorrent(filehash string) (*httpComm.TorrentPeersRsp, *httpComm.FailedRsp) {
	if len(filehash) == 49 {
		filehash = filehash[:46]
	}
	result, ontErr := sendRpcRequest("checktorrent", []interface{}{filehash})
	peers := &httpComm.TorrentPeersRsp{}
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, &httpComm.FailedRsp{ErrCode: berr.INVALID_PARAMS, ErrMsg: berr.ErrMap[berr.INVALID_PARAMS], FailedMsg: fmt.Sprintf("invalid fileHash len is %d, not 46", len(filehash))}
		case berr.INTERNAL_ERROR:
			return nil, &httpComm.FailedRsp{ErrCode: berr.INTERNAL_ERROR, ErrMsg: berr.ErrMap[berr.INTERNAL_ERROR], FailedMsg: ontErr.Error.Error()}
		}
		return nil, &httpComm.FailedRsp{ErrCode: ontErr.ErrorCode, ErrMsg: "", FailedMsg: ontErr.Error.Error()}
	}

	err := json.Unmarshal(result, peers)
	log.Debugf("cli.tracker.checktorrent success filehash: %s, peers: %v, err: %v\n", filehash, peers, err)

	if err != nil {
		return nil, &httpComm.FailedRsp{ErrCode: berr.JSON_UNMARSHAL_ERROR, ErrMsg: berr.ErrMap[berr.JSON_UNMARSHAL_ERROR], FailedMsg: err.Error()}
	}
	return peers, nil
}

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
