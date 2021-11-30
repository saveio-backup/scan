package restful

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/saveio/themis/common/log"
	"math"
	"strconv"

	chanCom "github.com/saveio/pylons/common"
	"github.com/saveio/pylons/transfer"
	httpComm "github.com/saveio/scan/http/base/common"
	"github.com/saveio/scan/http/base/error"
	"github.com/saveio/scan/http/base/rest"
	"github.com/saveio/scan/service"
	"github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/constants"
)

func GetFee(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)

	var cidStr string
	switch (params["ChannelID"]).(type) {
	case string:
		cidStr = params["ChannelID"].(string)
	}
	cid, err := strconv.Atoi(cidStr)
	if err != nil {
		cid = 0
		log.Error("get wrong channel id:", err)
	}

	fee, err := service.ScanNode.GetFee(uint64(cid))
	if err != nil {
		res["Desc"] = err
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	rsp := httpComm.ChannelFeeRsp{
		Flat:               uint64(fee.Flat),
		Proportional:       uint64(fee.Proportional),
		FlatFormat:         utils.FormatUsdt(uint64(fee.Flat)),
		ProportionalFormat: utils.FormatUsdt(uint64(fee.Proportional)),
	}
	res["Result"] = rsp
	return res
}

func GetChannelList(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)
	channelInfos, err := service.ScanNode.GetAllChannels()
	if err != nil {
		res["Desc"] = err
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	res["Result"] = channelInfos
	return res
}

func GetDeposit(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)
	var partnerAddrstr string
	switch (params["PartnerAddr"]).(type) {
	case string:
		partnerAddrstr = params["PartnerAddr"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	balance, err := service.ScanNode.QueryChannelDeposit(partnerAddrstr)
	if err != nil {
		res["Desc"] = err
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	rsp := httpComm.ChannelTotalDepositBalanceRsp{
		TotalDepositBalance: balance,
		TotalDepositBalanceFormat: utils.FormatUsdt(balance),
	}
	res["Result"] = rsp
	return res
}

func PostFee(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)

	var pwd string
	switch (params["Password"]).(type) {
	case string:
		pwd = params["Password"].(string)
	default:
		res["Error"] = error.PASSWORD_WRONG
		return res
	}
	b := verifyPassword(pwd)
	if !b {
		res["Error"] = error.PASSWORD_WRONG
		return res
	}

	var flatStr string
	var proStr string

	switch (params["FlatFormat"]).(type) {
	case string:
		flatStr = params["FlatFormat"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	switch (params["ProportionalFormat"]).(type) {
	case string:
		proStr = params["ProportionalFormat"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	flat, err := strconv.ParseFloat(flatStr, 10)
	if err != nil || flat < 0 || flat > 100000 {
		res["Desc"] = "FlatFormat range [0, 100000] "
		res["Error"] = error.INVALID_PARAMS
		return res
	}
	realFlat := uint64(flat * math.Pow10(constants.USDT_DECIMALS))

	pro, err := strconv.ParseFloat(proStr, 10)
	if err != nil || pro < 0 || pro > 1 {
		res["Desc"] = "ProportionalFormat range [0, 1] "
		res["Error"] = error.INVALID_PARAMS
		return res
	}
	realPro := uint64(pro * math.Pow10(constants.USDT_DECIMALS))

	fee := &transfer.FeeScheduleState{
		Flat:         chanCom.FeeAmount(realFlat),
		Proportional: chanCom.ProportionalFeeAmount(realPro),
	}
	err = service.ScanNode.SetFee(fee)
	if err != nil {
		res["Desc"] = err
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	return res
}

func PostChannelOpen(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)

	var pwd string
	switch (params["Password"]).(type) {
	case string:
		pwd = params["Password"].(string)
	default:
		res["Error"] = error.PASSWORD_WRONG
		return res
	}
	b := verifyPassword(pwd)
	if !b {
		res["Error"] = error.PASSWORD_WRONG
		return res
	}

	var partnerAddrstr string
	switch (params["PartnerAddr"]).(type) {
	case string:
		partnerAddrstr = params["PartnerAddr"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	id, err := service.ScanNode.OpenChannel(partnerAddrstr, 0)
	if err != nil {
		res["Desc"] = err
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
	switch (params["Password"]).(type) {
	case string:
		pwd = params["Password"].(string)
	default:
		res["Error"] = error.PASSWORD_WRONG
		return res
	}
	b := verifyPassword(pwd)
	if !b {
		res["Error"] = error.PASSWORD_WRONG
		return res
	}

	var partnerAddrstr string
	switch (params["PartnerAddr"]).(type) {
	case string:
		partnerAddrstr = params["PartnerAddr"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}
	err := service.ScanNode.CloseChannel(partnerAddrstr)
	if err != nil {
		res["Desc"] = err
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	return res
}

func PostDeposit(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)

	var pwd string
	switch (params["Password"]).(type) {
	case string:
		pwd = params["Password"].(string)
	default:
		res["Error"] = error.PASSWORD_WRONG
		return res
	}
	b := verifyPassword(pwd)
	if !b {
		res["Error"] = error.PASSWORD_WRONG
		return res
	}

	var partnerAddrstr string
	var depositStr string

	switch (params["PartnerAddr"]).(type) {
	case string:
		partnerAddrstr = params["PartnerAddr"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	switch (params["Deposit"]).(type) {
	case string:
		depositStr = params["Deposit"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	deposit, err := strconv.ParseFloat(depositStr, 10)
	if err != nil || deposit < 0 {
		res["Error"] = error.INVALID_PARAMS
		return res
	}
	realDeposit := uint64(deposit * math.Pow10(constants.USDT_DECIMALS))

	balance, err := service.ScanNode.QueryChannelDeposit(partnerAddrstr)
	if err != nil {
		res["Desc"] = err
		res["Error"] = error.CHANNEL_ERROR
		return res
	}

	realDeposit += balance
	err = service.ScanNode.DepositToChannel(partnerAddrstr, realDeposit)
	if err != nil {
		res["Desc"] = err
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	return res
}

func PostWithdraw(params map[string]interface{}) map[string]interface{} {
	res := rest.ResponsePack(error.SUCCESS)

	var pwd string
	switch (params["Password"]).(type) {
	case string:
		pwd = params["Password"].(string)
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
	var amountStr string

	switch (params["PartnerAddr"]).(type) {
	case string:
		partnerAddrstr = params["PartnerAddr"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	switch (params["Amount"]).(type) {
	case string:
		amountStr = params["Amount"].(string)
	default:
		res["Error"] = error.INVALID_PARAMS
		return res
	}

	_, err := service.ScanNode.QueryHostInfo(partnerAddrstr)
	if err != nil {
		res["Desc"] = err
		res["Error"] = error.CHANNEL_ERROR
		return res
	}

	amount, err := strconv.ParseFloat(amountStr, 10)
	if err != nil || amount < 0 {
		res["Desc"] = "Amount range [0, TotalDeposit] "
		res["Error"] = error.INVALID_PARAMS
		return res
	}
	realAmount := uint64(amount * math.Pow10(constants.USDT_DECIMALS))

	balance, err := service.ScanNode.QueryChannelWithdraw(partnerAddrstr)
	if err != nil {
		res["Desc"] = err
		res["Error"] = error.CHANNEL_ERROR
		return res
	}

	realAmount += balance
	err = service.ScanNode.WithdrawFromChannel(partnerAddrstr, realAmount)
	if err != nil {
		res["Desc"] = err
		res["Error"] = error.CHANNEL_ERROR
		return res
	}
	return res
}

func verifyPassword(pwd string) bool {
	pwdHash := Sha256HexStr(service.ScanNode.AccountPassword)
	if pwd != pwdHash {
		return false
	}
	return true
}

func Sha256HexStr(str string) string {
	pwdBuf := sha256.Sum256([]byte(str))
	pwdHash := hex.EncodeToString(pwdBuf[:])
	return pwdHash
}
