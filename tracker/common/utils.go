/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-03-19 
*/
package common

import (
	"bytes"
	netcomm "github.com/oniio/oniChain/common"
	)

func WHPTobyte(walletAddr,hostPort string)([]byte,[]byte){
	wAddr,err:=netcomm.AddressFromBase58(walletAddr)
	if err!=nil{
		return nil,nil
	}
	bf:=new(bytes.Buffer)
	if err=wAddr.Serialize(bf);err!=nil{
		return nil,nil
	}
	key:=bf.Bytes()
	value:=[]byte(hostPort)
	return key,value
}