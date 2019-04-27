package tracker

import (
	"crypto/rand"
	//"encoding/binary"
	"net"

	"encoding/json"
	"fmt"
	"github.com/anacrolix/dht/krpc"
	"github.com/oniio/oniChain/account"
	chainCom "github.com/oniio/oniChain/common"
	"github.com/oniio/oniChain/common/log"
	"github.com/oniio/oniChain/core/signature"
	"github.com/oniio/oniChain/core/types"
	"github.com/oniio/oniChain/crypto/keypair"
	"github.com/oniio/oniChain/errors"
	"github.com/oniio/oniDNS/cmd/utils"
	"github.com/oniio/oniDNS/common"
	pm "github.com/oniio/oniDNS/messages/protoMessages"
	"github.com/oniio/oniDNS/network/actor/recv"
	"github.com/oniio/oniDNS/storage"
	common2 "github.com/oniio/oniDNS/tracker/common"
)

// CompleteTorrent Complete make torrent
func CompleteTorrent(infoHash common.MetaInfoHash, trackerUrl string, nodeIP net.IP, port uint16) error {
	id := common.PeerID{}
	rand.Read(id[:])

	announce := Announce{
		TrackerUrl: trackerUrl,
		Request: AnnounceRequest{
			PeerId:    id,
			Left:      0,
			InfoHash:  infoHash,
			IPAddress: ipconvert(nodeIP),
			Port:      port,
			Event:     AnnounceEventCompleted,
		},
	}
	ret, err := announce.Do()
	if err != nil {
		log.Errorf("CompleteTorrent failed err:%s\n", err)
		return err
	}
	log.Debugf("interval:%d, leechers:%d, seeders:%d, peers:%v\n", ret.Interval, ret.Leechers, ret.Seeders, ret.Peers)
	return nil
}

// GetTorrentPeers get peers of torrent
func GetTorrentPeers(infoHash common.MetaInfoHash, trackerUrl string, numWant int32, left uint64) []Peer {
	id := common.PeerID{}
	rand.Read(id[:])
	announce := Announce{
		TrackerUrl: trackerUrl,
		Request: AnnounceRequest{
			PeerId:   id,
			Left:     left,
			InfoHash: infoHash,
			NumWant:  numWant,
		},
	}
	ret, err := announce.Do()
	if err != nil {
		log.Errorf("GetTorrentPeers failed err:%s\n", err)
		return nil
	}
	log.Debugf("interval:%d, leechers:%d, seeders:%d, peers:%v\n", ret.Interval, ret.Leechers, ret.Seeders, ret.Peers)
	return ret.Peers
}

// ---------Tracker client relative action------------
func RegEndPoint(trackerUrl string, sigData []byte, pubKey keypair.PublicKey, walletAddr [20]byte, nodeIP net.IP, port uint16) error {
	if err := common2.ParamVerify(trackerUrl, sigData, pubKey, walletAddr, nodeIP, port); err != nil {
		return err
	}
	id := common.PeerID{}
	rand.Read(id[:])
	announce := Announce{
		TrackerUrl: trackerUrl,
		Request: AnnounceRequest{
			PeerId:    id,
			IPAddress: ipconvert(nodeIP),
			Port:      port,
			Wallet:    walletAddr,
		},
		flag: ActionReg,
	}
	ret, err := announce.Do()
	if err != nil {
		log.Errorf("RegEndPoint failed err:%s\n", err)
		return err
	}
	log.Debugf("[RegEndPoint ]ip:%v, port:%d, wallet:%s\n", ret.IPAddress, ret.Port, ret.Wallet)
	return nil
}

func UnRegEndPoint(trackerUrl string, sigData []byte, pubKey keypair.PublicKey, walletAddr [20]byte) error {
	if err := common2.ParamVerify(trackerUrl, sigData, pubKey, walletAddr, nil, 0); err != nil {
		return err
	}
	id := common.PeerID{}
	rand.Read(id[:])
	announce := Announce{
		TrackerUrl: trackerUrl,
		Request: AnnounceRequest{
			PeerId: id,
			Wallet: walletAddr,
		},
		flag: ActionUnReg,
	}
	ret, err := announce.Do()
	if err != nil {
		log.Errorf("UnRegEndPoint failed err:%s\n", err)
		return err
	}
	log.Debugf("[UnRegEndPoint ]wallet:%s\n", ret.Wallet)
	return nil
}

func ReqEndPoint(trackerUrl string, walletAddr [20]byte) ([]byte, error) {
	id := common.PeerID{}
	rand.Read(id[:])
	announce := Announce{
		TrackerUrl: trackerUrl,
		Request: AnnounceRequest{
			PeerId: id,
			Wallet: walletAddr,
		},
		flag: ActionReq,
	}
	ret, err := announce.Do()
	if err != nil {
		log.Errorf("ReqEndPoint failed err:%s\n", err)
		return nil, err
	}
	var nodeAddr krpc.NodeAddr
	nodeAddr.IP = ret.IPAddress[:]
	nodeAddr.Port = int(ret.Port)
	nb, err := nodeAddr.MarshalBinary()
	if err != nil {
		return nil, err
	}
	log.Debugf("[ReqEndPoint ]wallet:%s,ip:%v, port:%d\n", ret.Wallet, ret.IPAddress, ret.Port)

	return nb, nil
}
func UpdateEndPoint(trackerUrl string, sigData []byte, pubKey keypair.PublicKey, walletAddr [20]byte, nodeIP net.IP, port uint16) error {
	if err := common2.ParamVerify(trackerUrl, sigData, pubKey, walletAddr, nodeIP, port); err != nil {
		return err
	}
	id := common.PeerID{}
	rand.Read(id[:])
	announce := Announce{
		TrackerUrl: trackerUrl,
		Request: AnnounceRequest{
			PeerId:    id,
			Wallet:    walletAddr,
			IPAddress: ipconvert(nodeIP),
			Port:      port,
		},
		flag: ActionUpdate,
	}
	ret, err := announce.Do()
	if err != nil {
		log.Errorf("RegEndPoint failed err:%s\n", err)
		return err
	}
	log.Debugf("[UpdateEndPoint]  wallet:%s, ip:%v, port:%d\n", ret.Wallet, ret.IPAddress, ret.Port)
	return nil
}

// ---------Local DDNS client relative action------------
//local endPointReg
func EndPointRegistry(walletAddr, hostPort string) error {
	log.Debugf("Local EndPointRegistry wallet:%s,host:%s\n", walletAddr, hostPort)
	if walletAddr == "" || hostPort == "" {
		return errors.NewErr("[EndPointRegistry] walletAddr or hostPort is null")
	}
	k, v := common.WHPTobyte(walletAddr, hostPort)
	if err := storage.TDB.Put(k, v); err != nil {
		return err
	}
	//hb, err := storage.TDB.Get(k)
	//if hb != nil && err == nil {
	//	log.Errorf("This wallet:%s had already registerd! Do not multiple registration", walletAddr)
	//	return nil
	//}
	m := &pm.Registry{
		WalletAddr: walletAddr,
		HostPort:   hostPort,
	}
	recv.P2pPid.Tell(m)
	return nil
}

//local endPointUpdate
func EndPointRegUpdate(walletAddr, hostPort string) error {
	if walletAddr == "" || hostPort == "" {
		return errors.NewErr("[EndPointRegistry] walletAddr or hostPort is null")
	}
	k, v := common.WHPTobyte(walletAddr, hostPort)
	exist, err := storage.TDB.Has(k)
	if !exist || err != nil {
		return errors.NewErr("[EndPointRegUpdate] wallet is not Registed!")
	}
	if err := storage.TDB.Put(k, v); err != nil {
		return err
	}
	m := &pm.Registry{
		WalletAddr: walletAddr,
		HostPort:   hostPort,
	}
	recv.P2pPid.Tell(m)
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
		return fmt.Errorf("[EndPointUnRegistry] wallet %s is not Registed!", walletAddr)
	}

	if err := storage.TDB.Delete(w); err != nil {
		return err
	}
	m := &pm.UnRegistry{
		WalletAddr: walletAddr,
	}
	recv.P2pPid.Tell(m)
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
		return "", fmt.Errorf("This wallet %s did not regist endpoint", walletAddr)
	}
	return string(hpBytes), nil
}
