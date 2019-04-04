package tracker

import (
	"crypto/rand"
	"encoding/binary"
	"net"

	"github.com/oniio/oniDNS/common"
	"github.com/oniio/oniChain/common/log"
	"github.com/oniio/oniChain/errors"
	"fmt"
	"github.com/oniio/oniDNS/storage"
	"github.com/oniio/oniDNS/messageBus"
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

func RegEndPoint(trackerUrl,walletAddr string,nodeIP net.IP, port uint16)error{
	id := common.PeerID{}
	rand.Read(id[:])
	announce := Announce{
		TrackerUrl: trackerUrl,
		Request:AnnounceRequest{
			IPAddress: binary.BigEndian.Uint32(nodeIP),
			Port:      port,
			Wallet:walletAddr,
		},
		flag:ActionReg,
	}
	ret, err := announce.Do()
	if err != nil {
		log.Errorf("GetTorrentPeers failed err:%s\n", err)
		return err
	}
	log.Debugf("interval:%d, leechers:%d, seeders:%d, peers:%v\n", ret.Interval, ret.Leechers, ret.Seeders, ret.Peers)
	return nil
}

//local endPointReg
func EndPointRegistry(walletAddr,hostPort string )error{
	if walletAddr==""|| hostPort==""{
		return errors.NewErr("[EndPointRegistry] walletAddr or hostPort is null")
	}
	k,v:=common.WHPTobyte(walletAddr,hostPort)
	if err:=storage.TDB.Put(k,v);err!=nil{
		return err
	}
	messageBus.MsgBus.MsgBox <- &messageBus.RegMsg{WalletAddr:walletAddr,HostPort:hostPort}
	return nil
}

func EndPointRegUpdate(walletAddr,hostPort string )error{
	if walletAddr==""|| hostPort==""{
		return errors.NewErr("[EndPointRegistry] walletAddr or hostPort is null")
	}
	k,v:=common.WHPTobyte(walletAddr,hostPort)
	exist,err:=storage.TDB.Has(k)
	if !exist || err!=nil{
		return errors.NewErr("[EndPointRegUpdate] wallet is not Registed!")
	}
	if err:=storage.TDB.Put(k,v);err!=nil{
		return err
	}
	messageBus.MsgBus.MsgBox <- &messageBus.RegMsg{WalletAddr:walletAddr,HostPort:hostPort}
	return nil
}

func EndPointUnRegistry(walletAddr string )error{
	if walletAddr==""{
		return errors.NewErr("[EndPointUnRegistry] walletAddr is null")
	}
	w,_:=common.WHPTobyte(walletAddr,"")
	exist,_:=storage.TDB.Has(w)
	if !exist{
		return fmt.Errorf("[EndPointUnRegistry] wallet %s is not Registed!",walletAddr)
	}

	if err:=storage.TDB.Delete(w);err!=nil{
		return err
	}
	messageBus.MsgBus.MsgBox <- &messageBus.RegMsg{WalletAddr:walletAddr}

	return nil
}

func EndPointQuerry(walletAddr string)(string,error){
	w,_:=common.WHPTobyte(walletAddr,"")
	hpBytes,err:=storage.TDB.Get(w)
	if err!=nil{
		return "",err
	}
	if hpBytes==nil{
		return "",fmt.Errorf("This wallet %s did not regist endpoint",walletAddr)
	}
	return string(hpBytes),nil
}

