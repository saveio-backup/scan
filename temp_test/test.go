/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-27
 */
package temp_test

func main() {
	endPoint := httpComm.EndPointRsp{
		Wallet: wAddr,
		Host:   host,
	}
	em, err := json.Marshal(endPoint)
	if err != nil {
		PrintErrorMsg("regEndPoint json.Marshal error:%s\n", err)
	}
	sigdata, err := utils.Sign(em, acc)
	err = utils.RegEndPoint(wAddr, host, sigdata, acc)
}

func verify() {
	//-----------for test---------------
	endPoint := httpComm.EndPointRsp{
		Wallet: wAddr,
		Host:   host,
	}
	em, _ := json.Marshal(endPoint)
	switch (params[2]).(type) {
	case []byte:
		sigData = params[2].([]byte)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	switch (params[3]).(type) {
	case *account.Account:
		acc = params[3].(*account.Account)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	if err := signature.Verify(acc.PublicKey, em, sigData); err != nil {
		return responsePack(berr.SIG_VERIFY_ERROR, "")
	}
	//-----------------for test end------------------------
}
