package tracker

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/anacrolix/dht/krpc"
	"github.com/saveio/scan/storage"
	tkComm "github.com/saveio/scan/tracker/common"
	themisComm "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/crypto/keypair"
)

type ActionEndpointRegParams struct {
	Wallet [20]byte
	IP     net.IP
	Port   uint16
}

type ActionTorrentCompleteParams struct {
	InfoHash storage.MetaInfoHash
	IP       net.IP
	Port     uint16
}

type ActionGetTorrentPeersParams struct {
	InfoHash storage.MetaInfoHash
	NumWant  int32
	Left     uint64
}

// CompleteTorrent Complete make torrent
func CompleteTorrent(trackerUrl string, params ActionTorrentCompleteParams, pubKey keypair.PublicKey, _sigCallback func(rawData []byte) ([]byte, error)) error {
	id := storage.PeerID{}
	rand.Read(id[:])

	rawData, err := json.Marshal(params)
	if err != nil {
		return err
	}
	sigData, err := _sigCallback(rawData)
	if err != nil {
		log.Infof("_sigCallback err %v", err)
		return err
	}

	fmt.Println("sigData", sigData, len(sigData))
	fmt.Println("pubKey", pubKey)
	fmt.Println("pubKey Binarys", keypair.SerializePublicKey(pubKey), len(keypair.SerializePublicKey(pubKey)))
	log.Debugf("CompleteTorrent Params %v\n", params)

	announce := Announce{
		TrackerUrl: trackerUrl,
		Request: AnnounceRequest{
			PeerId:    id,
			Left:      0,
			InfoHash:  params.InfoHash,
			IPAddress: ipconvert(params.IP),
			Port:      params.Port,
			Event:     AnnounceEventCompleted,
			NumWant:   1,
		},
		RequestOptions: AnnounceRequestOptions{
			PubKeyTLV:    tkComm.NewTLV(tkComm.TLV_TAG_PUBLIC_KEY, keypair.SerializePublicKey(pubKey)),
			SignatureTLV: tkComm.NewTLV(tkComm.TLV_TAG_SIGNATURE, sigData),
		},
		flag: ActionAnnounce,
	}
	log.Debugf("tracker.client.CompleteTorrent announce %v\n", announce)
	ret, err := announce.Do()

	if err != nil {
		log.Errorf("CompleteTorrent failed err: %s\n", err)
		return err
	}
	if len(ret.Peers) == 0 {
		return errors.New("no peers")
	}
	log.Debugf("tracker.client.CompleteTorrent interval:%d, leechers:%d, seeders:%d, peer:%s fileHash: %s\n", ret.Interval, ret.Leechers, ret.Seeders, krpc.NodeAddr{IP: params.IP, Port: int(params.Port)}.String(), string(params.InfoHash[:]))
	return nil
}

// GetTorrentPeers get peers of torrent
func GetTorrentPeers(trackerUrl string, params ActionGetTorrentPeersParams, pubKey keypair.PublicKey, _sigCallback func(rawData []byte) ([]byte, error)) ([]Peer, error) {
	id := storage.PeerID{}
	rand.Read(id[:])

	rawData, err := json.Marshal(params)
	if err != nil {
		return []Peer{}, err
	}
	fmt.Println("rawData", rawData)
	sigData, err := _sigCallback(rawData)
	if err != nil {
		log.Infof("_sigCallback err %v", err)
		return []Peer{}, err
	}

	fmt.Println("sigData", sigData, len(sigData))
	fmt.Println("pubKey", pubKey)
	fmt.Println("pubKey Binarys", keypair.SerializePublicKey(pubKey), len(keypair.SerializePublicKey(pubKey)))
	log.Debugf("GetTorrentPeers Params %v\n", params)

	announce := Announce{
		TrackerUrl: trackerUrl,
		Request: AnnounceRequest{
			PeerId:   id,
			Left:     params.Left,
			InfoHash: params.InfoHash,
			NumWant:  params.NumWant,
		},
		RequestOptions: AnnounceRequestOptions{
			PubKeyTLV:    tkComm.NewTLV(tkComm.TLV_TAG_PUBLIC_KEY, keypair.SerializePublicKey(pubKey)),
			SignatureTLV: tkComm.NewTLV(tkComm.TLV_TAG_SIGNATURE, sigData),
		},
		flag: ActionAnnounce,
	}
	log.Debugf("tracker.client.GetTorrentPeers announce %v\n", announce)
	ret, err := announce.Do()
	if err != nil {
		log.Errorf("GetTorrentPeers failed err: %s\n", err)
		return []Peer{}, nil
	}
	log.Debugf("tracker.client.CompleteTorrent interval:%d, leechers:%d, seeders:%d, peers:%v, infoHash:%v \n", ret.Interval, ret.Leechers, ret.Seeders, ret.Peers, string(params.InfoHash[:]))
	return ret.Peers, nil
}

func RegEndPoint(trackerUrl string, params ActionEndpointRegParams, pubKey keypair.PublicKey, _sigCallback func(rawData []byte) ([]byte, error)) error {
	rawData, err := json.Marshal(params)
	if err != nil {
		return err
	}
	fmt.Println("rawData", rawData)
	sigData, err := _sigCallback(rawData)
	if err != nil {
		log.Infof("_sigCallback err %v", err)
		return err
	}

	fmt.Println(params)
	fmt.Println("sigData", sigData, len(sigData))
	fmt.Println("pubKey", pubKey)
	fmt.Println("pubKey Binarys", keypair.SerializePublicKey(pubKey), len(keypair.SerializePublicKey(pubKey)))
	log.Debugf("RegEndPoint Params %v\n", params)

	id := storage.PeerID{}
	rand.Read(id[:])
	announce := Announce{
		TrackerUrl: trackerUrl,
		Request: AnnounceRequest{
			PeerId:    id,
			IPAddress: ipconvert(params.IP),
			Port:      params.Port,
			Wallet:    params.Wallet,
		},
		RequestOptions: AnnounceRequestOptions{
			PubKeyTLV:    tkComm.NewTLV(tkComm.TLV_TAG_PUBLIC_KEY, keypair.SerializePublicKey(pubKey)),
			SignatureTLV: tkComm.NewTLV(tkComm.TLV_TAG_SIGNATURE, sigData),
		},
		flag: ActionReg,
	}
	ret, err := announce.Do()
	if err != nil {
		log.Errorf("RegEndPoint failed err: %s\n", err)
		return err
	}
	hostIP := net.IP(ret.IPAddress[:])
	addr, err := themisComm.AddressParseFromBytes(ret.Wallet[:])
	if err != nil {
		log.Debugf("decode ret.Wallet err: %v\n", err)
	}
	log.Infof("tracker client [RegEndPoint] wallet:%s, nodeAddr %s:%d\n", addr.ToBase58(), hostIP.String(), ret.Port)
	return nil
}

func ReqEndPoint(trackerUrl string, walletAddr [20]byte) (string, error) {
	id := storage.PeerID{}
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
	log.Debugf("tracker.client.ReqEndpoint ret: %v, err: %v\n", ret, err)
	hostIP := net.IP(ret.IPAddress[:])
	if hostIP.String() == "0.0.0.0" || int(ret.Port) == 0 {
		return "", errors.New(fmt.Sprintf("endpoint host or port is 0, nodeAddr %s:%d", hostIP.String(), ret.Port))
	}
	addr, err := themisComm.AddressParseFromBytes(ret.Wallet[:])
	if err != nil {
		log.Debugf("decode ret.Wallet err: %v\n", err)
	}
	log.Infof("tracker client [ReqEndPoint] wallet:%s, nodeAddr %s:%d\n", addr.ToBase58(), hostIP.String(), ret.Port)
	return fmt.Sprintf("%s:%d", hostIP.String(), ret.Port), nil
}

func RegNodeType(trackerUrl string, walletAddr [20]byte, nodeIP net.IP, port uint16, nodeType NodeType) error {
	id := storage.PeerID{}
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
	addr, err := themisComm.AddressParseFromBytes(ret.Wallet[:])
	if err != nil {
		log.Debugf("decode ret.Wallet err: %v\n", err)
	}
	log.Debugf("[UpdateEndPoint]  wallet:%s, ip:%v, port:%d\n", addr.ToBase58(), ret.IPAddress, ret.Port)
	return nil
}

func GetNodesByType(trackerUrl string, nodeType NodeType) (*NodesInfoSt, error) {
	id := storage.PeerID{}
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
	addr, err := themisComm.AddressParseFromBytes(ret.Wallet[:])
	if err != nil {
		log.Debugf("decode ret.Wallet err: %v\n", err)
	}
	log.Debugf("[UpdateEndPoint]  wallet:%s, ip:%v, port:%d\n", addr.ToBase58(), ret.IPAddress, ret.Port)
	return ret.NodesInfo, nil
}
