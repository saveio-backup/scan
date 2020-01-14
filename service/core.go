package service

import (
	theComm "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	fs "github.com/saveio/themis/smartcontract/service/native/savefs"
)

func (this *Node) GetCurrentBlockHeight() (uint32, error) {
	height, err := this.Chain.GetCurrentBlockHeight()
	if err != nil {
		return 0, err
	}
	return height, nil
}

func (this *Node) GetWhiteList(fileHash string) (*fs.WhiteList, error) {
	list, err := this.Chain.Native.Fs.GetWhiteList(fileHash)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (this *Node) InWhiteListOrNot(fileHash string, wallAddr theComm.Address) (bool, error) {
	wl, err := this.GetWhiteList(fileHash)
	log.Debugf("whitelist: %v, err: %v", wl, err)
	if err != nil {
		return true, err
	}
	height, err := this.GetCurrentBlockHeight()
	log.Debugf("height: %d", height)
	if err != nil {
		return true, err
	}
	for _, rule := range wl.List {
		log.Debugf("rule: %v, addr: %v, wall: %v", rule, rule.Addr.ToBase58(), wallAddr.ToBase58())
		if rule.BaseHeight <= uint64(height) &&
			rule.ExpireHeight > uint64(height) &&
			rule.Addr.ToBase58() == wallAddr.ToBase58() {
			return true, nil
		}
	}
	return false, nil
}
