package tracker

import (
	"crypto/rand"
	"encoding/binary"
	"net"

	"github.com/oniio/oniDNS/common"
	cm "github.com/oniio/oniDNS/tracker/common"
	"github.com/oniio/oniChain/common/log"
	"github.com/oniio/oniChain/errors"
	"fmt"
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
			IPAddress: binary.BigEndian.Uint32(nodeIP),
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

func EndPointRegistry(walletAddr,hostPort string )error{
	if walletAddr==""|| hostPort==""{
		return errors.NewErr("[EndPointRegistry] walletAddr or hostPort is null")
	}
	if err:=RegMsgDB.Put(walletAddr,hostPort);err!=nil{
		return err
	}
	return nil
}

func EndPointUnRegistry(walletAddr string )error{
	if walletAddr==""{
		return errors.NewErr("[EndPointUnRegistry] walletAddr is null")
	}
	w,_:=cm.WHPTobyte(walletAddr,"")
	exist,_:=RegMsgDB.Has(w)
	if !exist{
		return errors.NewErr("[EndPointUnRegistry] wallet is not Registed!")
	}

	if err:=RegMsgDB.Delete(walletAddr);err!=nil{
		return err
	}
	return nil
}

func EndPointQuerry(walletAddr string)(string,error){
	hpBytes,err:=RegMsgDB.Get(walletAddr)
	if err!=nil{
		return "",err
	}
	if hpBytes==nil{
		return "",fmt.Errorf("This wallet %s did not regist endpoint",walletAddr)
	}
	return string(hpBytes),nil
}

