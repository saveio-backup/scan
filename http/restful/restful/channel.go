package restful

import (
	"fmt"
	chanCom "github.com/saveio/pylons/common"
	"github.com/saveio/pylons/transfer"
	httpComm "github.com/saveio/scan/http/base/common"
	"github.com/saveio/scan/http/base/error"
	"github.com/saveio/scan/http/base/rest"
	"github.com/saveio/scan/service"
)

func GetFee(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)
	fee, err := service.ScanNode.GetFee()
	if err != nil {
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	rsp := httpComm.ChannelFeeRsp{
		Flat:         uint64(fee.Flat),
		Proportional: uint64(fee.Proportional),
	}
	res["Result"] = rsp
	return res
}

func GetChannelList(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)
	channelInfos, err := service.ScanNode.GetAllChannels()
	if err != nil {
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	res["Result"] = channelInfos
	return res
}

func GetDeposit(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)
	var partnerAddrstr string
	switch (params["partnerAddr"]).(type) {
	case string:
		partnerAddrstr = params["partnerAddr"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	balance, err := service.ScanNode.QuerySpecialChannelDeposit(partnerAddrstr)
	if err != nil {
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	rsp := httpComm.ChannelCurrentBalanceRsp{
		CurrentBalance: balance,
	}
	res["Result"] = rsp
	return res
}

func PostFee(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)

	var pwd string
	switch (params["password"]).(type) {
	case string:
		pwd = params["password"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}
	b := verifyPassword(pwd)
	if !b {
		res["Error"] = error.PASSWORD_WRONG
		return res
	}

	var flat uint64
	var pro uint64

	switch (params["flat"]).(type) {
	case float64:
		flat = uint64(params["flat"].(float64))
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	switch (params["proportional"]).(type) {
	case float64:
		pro = uint64(params["proportional"].(float64))
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	fee := &transfer.FeeScheduleState{
		Flat: chanCom.FeeAmount(flat),
		Proportional: chanCom.ProportionalFeeAmount(pro),
	}
	err := service.ScanNode.SetFee(fee)
	if err != nil {
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	return res
}

func PostChannelOpen(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)

	var pwd string
	switch (params["password"]).(type) {
	case string:
		pwd = params["password"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}
	b := verifyPassword(pwd)
	if !b {
		res["Error"] = error.PASSWORD_WRONG
		return res
	}

	var partnerAddrstr string
	switch (params["partnerAddr"]).(type) {
	case string:
		partnerAddrstr = params["partnerAddr"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	id, err := service.ScanNode.OpenChannel(partnerAddrstr, 0)
	if err != nil {
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	rsp := httpComm.ChannelRsp{
		Id: uint32(id),
	}
	res["Result"] = rsp
	return res
}

func Postchannelclose(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)

	var pwd string
	switch (params["password"]).(type) {
	case string:
		pwd = params["password"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}
	b := verifyPassword(pwd)
	if !b {
		res["Error"] = error.PASSWORD_WRONG
		return res
	}

	var partnerAddrstr string
	switch (params["partnerAddr"]).(type) {
	case string:
		partnerAddrstr = params["partnerAddr"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}
	err := service.ScanNode.CloseChannel(partnerAddrstr)
	if err != nil {
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	return res
}

func PostDeposit(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)

	var pwd string
	switch (params["password"]).(type) {
	case string:
		pwd = params["password"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}
	b := verifyPassword(pwd)
	if !b {
		res["Error"] = error.PASSWORD_WRONG
		return res
	}

	var partnerAddrstr string
	var totalDeposituint64 uint64
	switch (params["partnerAddr"]).(type) {
	case string:
		partnerAddrstr = params["partnerAddr"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	switch (params["totalDeposit"]).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		totalDeposituint64 = params["totalDeposit"].(uint64)
	case float32, float64:
		// may be bugs
		totalDeposituint64 = uint64(params["totalDeposit"].(float64))
		fmt.Println(totalDeposituint64)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	err := service.ScanNode.DepositToChannel(partnerAddrstr, totalDeposituint64)
	if err != nil {
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	return res
}

func PostWithdraw(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)

	var pwd string
	switch (params["password"]).(type) {
	case string:
		pwd = params["password"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}
	b := verifyPassword(pwd)
	if !b {
		res["Error"] = error.PASSWORD_WRONG
		return res
	}

	var partnerAddrstr string
	var amountuint64 uint64
	switch (params["partnerAddr"]).(type) {
	case string:
		partnerAddrstr = params["partnerAddr"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	switch (params["amount"]).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		amountuint64 = params["amount"].(uint64)
	case float32, float64:
		// may be bugs
		amountuint64 = uint64(params["amount"].(float64))
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	host, err := service.ScanNode.QueryHostInfo(partnerAddrstr)
	if host == "" {
		res["Error"] = error.CHANNEL_ERROR
		return res
	}

	err = service.ScanNode.ChannelWithdraw(partnerAddrstr, amountuint64)
	if err != nil {
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	return res
}

func verifyPassword(pwd string) bool {
	if pwd != service.ScanNode.AccountPassword {
		return false
	}
	return true
}