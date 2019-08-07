package tracker

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/anacrolix/dht/krpc"
	"github.com/saveio/scan/channel"
	"github.com/saveio/scan/common"
	"github.com/saveio/scan/common/config"
	pm "github.com/saveio/scan/messages/protoMessages"
	"github.com/saveio/scan/network"
	"github.com/saveio/scan/storage"

	chaincom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
)

// ---------Local DDNS client relative action------------
//local endPointReg
func EndPointRegistry(walletAddr, hostPort string) error {
	log.Debugf("Local EndPointRegistry wallet:%s,host:%s\n", walletAddr, hostPort)
	if walletAddr == "" || hostPort == "" {
		return errors.NewErr("[EndPointRegistry] walletAddr or hostPort is null")
	}

	host, port := strings.Split(hostPort, ":")[0], strings.Split(hostPort, ":")[1]

	wb, _ := chaincom.AddressFromBase58(walletAddr)

	portint, err := strconv.Atoi(port)
	// netIp := net.ParseIP(host).To4()
	// if netIp == nil {
	// 	netIp = net.ParseIP(host).To16()
	// }
	nodeAddr := krpc.NodeAddr{
		IP:   net.ParseIP(host).To4(),
		Port: portint,
	}
	nb, err := nodeAddr.MarshalBinary()
	bf := new(bytes.Buffer)
	if err = wb.Serialize(bf); err != nil {
		return err
	}
	key := bf.Bytes()
	// k, v := common.WHPTobyte(walletAddr, hostPort)
	if err := storage.TDB.Put(key, nb); err != nil {
		return err
	}
	// hb, err := storage.TDB.Get(k)
	// if hb != nil && err == nil {
	// 	log.Errorf("This wallet:%s had already registerd! Do not multiple registration", walletAddr)
	// 	return nil
	// }
	channel.GlbChannelSvr.Channel.SetHostAddr(walletAddr, fmt.Sprintf("%s://%s", config.Parameters.Base.ChannelProtocol, fmt.Sprintf("%s:%s", host, port)))
	m := &pm.Registry{
		WalletAddr: walletAddr,
		HostPort:   fmt.Sprintf("%s:%s", host, port),
		Type:       0,
	}
	network.DDNSP2P.GetPID().Tell(m)
	return nil
}

//local endPointUpdate
func EndPointRegUpdate(walletAddr, hostPort string) error {
	if walletAddr == "" || hostPort == "" {
		return errors.NewErr("[EndPointRegUpdate] walletAddr or hostPort is null")
	}
	k, v := common.WHPTobyte(walletAddr, hostPort)
	exist, err := storage.TDB.Has(k)
	if !exist || err != nil {
		return errors.NewErr("not exist")
	}
	if err := storage.TDB.Put(k, v); err != nil {
		return err
	}
	m := &pm.Registry{
		WalletAddr: walletAddr,
		HostPort:   hostPort,
		Type:       0,
	}
	network.DDNSP2P.GetPID().Tell(m)
	return nil
}

//local endPointUnReg
func EndPointUnRegistry(walletAddr string) error {
	if walletAddr == "" {
		return errors.NewErr("[EndPointUnRegistry] walletAddr is null")
	}
	w, _ := common.WHPTobyte(walletAddr, "")
	exist, _ := storage.TDB.Has(w)
	if !exist {
		return fmt.Errorf("not exist")
	}

	if err := storage.TDB.Delete(w); err != nil {
		return err
	}
	m := &pm.UnRegistry{
		WalletAddr: walletAddr,
		Type:       0,
	}
	network.DDNSP2P.GetPID().Tell(m)
	return nil
}

//local endPointQuery
func EndPointQuery(walletAddr string) (string, error) {
	w, _ := common.WHPTobyte(walletAddr, "")
	hpBytes, err := storage.TDB.Get(w)
	if err != nil {
		return "", err
	}
	if hpBytes == nil {
		return "", fmt.Errorf("not found")
	}
	var nodeAddr krpc.NodeAddr
	log.Infof("EndPointQuery: %v, hpBytes: %v", walletAddr, hpBytes)
	nodeAddr.UnmarshalBinary(hpBytes)
	log.Infof("EndPointQuery nodeAddr: %s:%d", nodeAddr.IP, nodeAddr.Port)

	return fmt.Sprintf("%s:%d", nodeAddr.IP, nodeAddr.Port), nil
}
