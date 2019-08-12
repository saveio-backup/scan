package tracker

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"

	"github.com/saveio/scan/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/crypto/keypair"
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
		flag: ActionAnnounce,
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
		flag: ActionAnnounce,
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
	log.Debugf("RegEndPoint Params: trackerUrl: %s, sigData: %v, pubKey: %v, WalletAddr: %v, nodeIP: %v, Port: %d", trackerUrl, sigData, pubKey, walletAddr, nodeIP, port)
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
	hostIP := net.IP(ret.IPAddress[:])
	log.Infof("tracker client [RegEndPoint] wallet:%s, nodeAddr %s:%d\n", ret.Wallet, hostIP.String(), ret.Port)
	return nil
}

func UnRegEndPoint(trackerUrl string, walletAddr [20]byte) error {
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
	log.Infof("tracker client [UnRegEndPoint] wallet:%s\n", ret.Wallet)
	return nil
}

func ReqEndPoint(trackerUrl string, walletAddr [20]byte) (string, error) {
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
		log.Errorf("ReqEndPoint err: %s\n", err)
		return "", err
	}
	hostIP := net.IP(ret.IPAddress[:])
	if hostIP.String() == "0.0.0.0" || int(ret.Port) == 0 {
		return "", errors.New(fmt.Sprintf("endpoint host or port is 0, nodeAddr %s:%d", hostIP.String(), ret.Port))
	}
	log.Infof("tracker client [ReqEndPoint] wallet:%s, nodeAddr %s:%d\n", ret.Wallet, hostIP.String(), ret.Port)
	return fmt.Sprintf("%s:%d", hostIP.String(), ret.Port), nil
}

func RegNodeType(trackerUrl string, walletAddr [20]byte, nodeIP net.IP, port uint16, nodeType NodeType) error {
	id := common.PeerID{}
	rand.Read(id[:])
	announce := Announce{
		TrackerUrl: trackerUrl,
		Request: AnnounceRequest{
			PeerId:    id,
			Wallet:    walletAddr,
			IPAddress: ipconvert(nodeIP),
			Port:      port,
			NodeType:  nodeType,
		},
		flag: ActionRegNodeType,
	}
	ret, err := announce.Do()
	if err != nil {
		log.Errorf("RegEndPoint failed err:%s\n", err)
		return err
	}
	log.Debugf("[UpdateEndPoint]  wallet:%s, ip:%v, port:%d\n", ret.Wallet, ret.IPAddress, ret.Port)
	return nil
}

func GetNodesByType(trackerUrl string, nodeType NodeType) (*NodesInfoSt, error) {
	id := common.PeerID{}
	rand.Read(id[:])
	announce := Announce{
		TrackerUrl: trackerUrl,
		Request: AnnounceRequest{
			PeerId:   id,
			NodeType: nodeType,
		},
		flag: ActionGetNodesByType,
	}

	ret, err := announce.Do()
	if err != nil {
		log.Errorf("RegEndPoint failed err:%s\n", err)
		return nil, err
	}
	log.Debugf("[UpdateEndPoint]  wallet:%s, ip:%v, port:%d\n", ret.Wallet, ret.IPAddress, ret.Port)
	return ret.NodesInfo, nil
}
