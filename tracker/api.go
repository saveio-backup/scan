package tracker

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	pm "github.com/saveio/scan/messages/protoMessages"
	network "github.com/saveio/scan/p2p/networks/dns"
	"github.com/saveio/scan/storage"

	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
)

// ---------Local DDNS client relative action------------
func CheckTorrent(fileHash string) ([]Peer, error) {
	if len(fileHash) != 46 {
		return nil, errors.NewErr(fmt.Sprintf("invalid fileHash len is %d, not 46", len(fileHash)))
	}

	v, err := storage.TDB.Get([]byte(fileHash))
	log.Infof("tracker.server.getTorrent: isNil %v, err %v", v == nil, err)
	if v == nil || err != nil {
		return []Peer{}, err
	}

	var t torrent
	json.Unmarshal(v, &t)
	log.Infof("tracker.server.getTorrent: %v", t)

	var peers []Peer
	for _, p := range t.Peers {
		peers = append(peers, Peer{
			IP:   p.IP[:],
			Port: int(p.Port),
		})
	}
	return peers, nil
}

//local endPointReg
func EndPointRegistry(walletAddr, hostAddr string) error {
	log.Debugf("Local EndPointRegistry wallet:%s,host:%s\n", walletAddr, hostAddr)
	if walletAddr == "" || hostAddr == "" {
		return errors.NewErr("[EndPointRegistry] walletAddr or hostPort is null")
	}
	index := strings.Index(hostAddr, "://")
	hostPort := hostAddr
	if index != -1 {
		hostPort = hostAddr[index+3:]
	}
	err := PutEndpoint(walletAddr, hostAddr)
	if err != nil {
		return err
	}
	network.DnsP2p.GetPID().Tell(&pm.Registry{WalletAddr: walletAddr, HostPort: hostPort, Type: 0})
	return nil
}

//local endPointUpdate
func EndPointRegUpdate(walletAddr, hostAddr string) error {
	if walletAddr == "" || hostAddr == "" {
		return errors.NewErr("[EndPointRegUpdate] walletAddr or hostPort is null")
	}
	index := strings.Index(hostAddr, "://")
	hostPort := hostAddr
	if index != -1 {
		hostPort = hostAddr[index+3:]
	}
	err := PutEndpoint(walletAddr, hostAddr)
	if err != nil {
		return err
	}

	network.DnsP2p.GetPID().Tell(&pm.Registry{WalletAddr: walletAddr, HostPort: hostPort, Type: 0})
	return nil
}

//local endPointUnReg
func EndPointUnRegistry(walletAddr string) error {
	err := storage.EDB.DelEndpoint(walletAddr)
	if err != nil {
		return err
	}
	m := &pm.UnRegistry{WalletAddr: walletAddr, Type: 0}
	network.DnsP2p.GetPID().Tell(m)
	return nil
}

//local endPointQuery
func EndPointQuery(walletAddr string) (string, error) {
	nodeAddr, err := storage.EDB.GetEndpoint(walletAddr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", nodeAddr.NodeAddr), nil
}

func PutEndpoint(walletAddr, hostAddr string) error {
	index := strings.Index(hostAddr, "://")
	hostPort := hostAddr
	if index != -1 {
		hostPort = hostAddr[index+3:]
	}
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return err
	}
	netIp := net.ParseIP(host).To4()
	if netIp == nil {
		netIp = net.ParseIP(host).To16()
	}
	netPort, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	err = storage.EDB.PutEndpoint(walletAddr, netIp, netPort)
	if err != nil {
		return err
	}
	return nil
}
