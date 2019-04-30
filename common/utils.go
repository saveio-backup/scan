/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-03-19
 */
package common

import (
	"bytes"
	netcomm "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"os"
	"os/signal"
	"syscall"
)

func WHPTobyte(walletAddr, hostPort string) ([]byte, []byte) {
	wAddr, err := netcomm.AddressFromBase58(walletAddr)
	if err != nil {
		return nil, nil
	}
	bf := new(bytes.Buffer)
	if err = wAddr.Serialize(bf); err != nil {
		return nil, nil
	}
	key := bf.Bytes()
	value := []byte(hostPort)
	return key, value
}

func WaitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("seeds received exit signal:%v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}
