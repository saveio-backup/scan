/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-15 
*/
package utils

import (
	"fmt"
	//"github.com/oniio/oniDNS/http/base/common"
	//"encoding/json"
	"github.com/oniio/oniChain/common/log"
	"github.com/oniio/oniChain/account"
	httpComm "github.com/oniio/oniDNS/http/base/common"
	"encoding/json"
)

func RegEndPoint(waddr,host string,sigData []byte,a *account.Account)error{
	result, ontErr := sendRpcRequest("regendpoint", []interface{}{waddr,host,sigData,a})
	if ontErr != nil {
		switch ontErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return fmt.Errorf("invalid address:%s,host:%s", waddr,host)
		}
		return ontErr.Error
	}
	endPoint:=&httpComm.EndPointRsp{}

	err:=json.Unmarshal(result,endPoint)
	if err != nil {
		return fmt.Errorf("json.Unmarshal error:%s", err)
	}
	log.Infof("DDNS endpoint registed success wallet:%s,host%s",endPoint.Wallet,endPoint.Host)
	log.Debugf("RegEndPoint result :%s",result)
	return nil
}