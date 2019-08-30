package tracker

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	pm "github.com/saveio/scan/p2p/actor/messages"
	network "github.com/saveio/scan/p2p/networks/dns"
	"github.com/saveio/scan/storage"

	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
)

// ---------Local DDNS client relative action------------
func CheckTorrent(fileHash string) ([]*storage.PeerInfo, error) {
	if len(fileHash) != 46 {
		return nil, errors.NewErr(fmt.Sprintf("invalid fileHash len is %d, not 46", len(fileHash)))
	}

	torrentPeers, _, _, err := storage.TDB.GetTorrentPeersByFileHash([]byte(fileHash), -1)
	return torrentPeers, err
}

//local endPointReg
func EndPointRegistry(walletAddr, hostAddr string) error {
	log.Debugf("Local EndPointRegistry walletAddr:%s, hostAddr:%s\n", walletAddr, hostAddr)
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
	network.DnsP2p.GetPID().Tell(&pm.Endpoint{WalletAddr: walletAddr, HostPort: hostPort, Type: 0})
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

	network.DnsP2p.GetPID().Tell(&pm.Endpoint{WalletAddr: walletAddr, HostPort: hostPort, Type: 0})
	return nil
}

//local endPointUnReg
func EndPointUnRegistry(walletAddr string) error {
	err := storage.EDB.DelEndpoint(walletAddr)
	if err != nil {
		return err
	}
	return nil
}

//local endPointQuery
func EndPointQuery(walletAddr string) (string, error) {
	nodeAddr, err := storage.EDB.GetEndpoint(walletAddr)
	if err != nil {
		return "", err
	}
	if nodeAddr == nil {
		return "", errors.NewErr("endpoint nil")
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
