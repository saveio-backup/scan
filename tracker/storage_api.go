/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-03-14 
*/
package tracker

import (
	"github.com/oniio/oniChain/common/log"
	"github.com/oniio/oniChain/errors"
	"fmt"
	"github.com/oniio/oniDNS/storage"
	"github.com/oniio/oniDNS/tracker/common"
)

var RegMsgDB *RegMsgStorage

type RegMsgStorage struct {
	Ls *storage.LevelDBStore
	*Network
}

func InitRegMsgDB(path string)(*RegMsgStorage,error){
	db,err:=NewStorage(path)
	if err!=nil{
		return nil,err
	}
	return db,nil
}

func NewStorage(path string) (*RegMsgStorage,error) {
	nls, err := storage.NewLevelDBStore(path)
	if err != nil {
		log.Errorf("init torrent cache err:%s\n", err)
		return nil,err
	}
	return &RegMsgStorage{Ls:nls},nil
}

//Put a key-value pair to leveldb
func (self *RegMsgStorage) Put(key,value string) error {
	if err:=self.SyncRegMsg(key,value);err!=nil{
		return fmt.Errorf("[Put] SyncRegMsg error:%v",err)
	}
	k,v:=common.WHPTobyte(key,value)
	if k==nil||v==nil{
		return errors.NewErr("[StrTobyte] syncRegMsg put,wallet or hostport is nil")
	}
	if err:=self.Ls.Put(k, v);err!=nil{
		return err
	}
	return nil
}

//Get the value of a key from leveldb
func (self *RegMsgStorage) Get(key string) ([]byte, error) {
	k,_:=common.WHPTobyte(key,"")
	return self.Ls.Get(k)

}

//Has return whether the key is exist in leveldb
func (self *RegMsgStorage) Has(key []byte) (bool, error) {
	return self.Ls.Has(key)
}

//Delete the the in leveldb
func (self *RegMsgStorage) Delete(key string) error {
	if err:=self.SyncUnRegMsg(key);err!=nil{
		return fmt.Errorf("[Delete] SyncUnRegMsg error:%v",err)
	}
	k,_:=common.WHPTobyte(key,"")
	return self.Ls.Delete(k)
}
