package storage

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/anacrolix/dht/krpc"
	"github.com/saveio/scan/common/config"
	"github.com/saveio/themis/common/log"
)

var TDB *TorrentDB

const (
	MAX_PEERS_LENGTH                 = 100
	DEFAULT_NUMWANT_TO_GET_ALL_PEERS = -1
)

type PeerID [20]byte

type TorrentDB struct {
	db *LevelDBStore
}

type PeerInfo struct {
	ID        PeerID
	Complete  bool
	IP        [4]byte
	Port      uint16
	NodeAddr  krpc.NodeAddr
	Timestamp time.Time
}

type Torrent struct {
	Leechers int32
	Seeders  int32
	Peers    map[string]*PeerInfo
}

func NewTorrentDB(db *LevelDBStore) *TorrentDB {
	return &TorrentDB{
		db: db,
	}
}

func (this *TorrentDB) Close() error {
	return this.db.Close()
}

func (this *TorrentDB) PutTorrent(fileHash []byte, torrent *Torrent) error {
	bt, err := json.Marshal(torrent)
	if err != nil {
		return err
	}
	return this.db.Put(fileHash, bt)
}

func (this *TorrentDB) GetTorrent(fileHash []byte) (*Torrent, error) {
	var t Torrent
	v, err := this.db.Get(fileHash)
	json.Unmarshal(v, &t)
	if v == nil || err != nil {
		return nil, err
	}
	return &t, nil
}

func (this *TorrentDB) GetTorrentBinary(fileHash []byte) (bt []byte, err error) {
	bt, err = this.db.Get(fileHash)
	if err != nil {
		return nil, err
	}
	return bt, nil
}

func (this *TorrentDB) DelTorrent(fileHash []byte) error {
	return this.db.Delete(fileHash)
}

func (this *TorrentDB) AddTorrentPeer(fileHash []byte, left uint64, nodeAddr string, peer *PeerInfo) error {
	t, err := this.GetTorrent(fileHash)
	if err != nil && err.Error() != "not found" {
		return err
	}

	if t == nil {
		t = &Torrent{}
	}

	if !peer.Complete && left == 0 {
		t.Leechers -= 1
		t.Seeders += 1
		peer.Complete = true
	} else if left == 0 {
		t.Seeders++
	} else {
		t.Leechers++
	}

	if t.Peers == nil {
		t.Peers = make(map[string]*PeerInfo, 0)
	}
	peer.Timestamp = time.Now()

	t.Peers[nodeAddr] = peer
	return this.PutTorrent(fileHash, t)
}

func (this *TorrentDB) DelTorrentPeer(fileHash []byte, peer *PeerInfo) error {
	t, err := this.GetTorrent(fileHash)
	if err != nil {
		return err
	}

	if t == nil {
		return nil
	}

	if peer.Complete {
		t.Seeders -= 1
	} else {
		t.Leechers -= 1
	}
	delete(t.Peers, peer.NodeAddr.String())
	return this.PutTorrent(fileHash, t)
}

func (this *TorrentDB) GetTorrentPeersByFileHash(fileHash []byte, numWant int32) (retPeers []*PeerInfo, leechers, seeders int32, err error) {
	t, err := this.GetTorrent(fileHash)
	if err != nil {
		return retPeers, leechers, seeders, err
	}
	if t == nil {
		return retPeers, leechers, seeders, errors.New("torrent not found, got nil")
	}
	leechers = t.Leechers
	seeders = t.Seeders
	log.Debugf("torrent %v\n", t)

	peersCount, peersNotExpireCount := 0, 0
	peersNotExpire := make(map[string]*PeerInfo)
	now := time.Now()
	expireDuration, err := time.ParseDuration(config.Parameters.Base.TrackerPeerValidDuration)
	if err != nil {
		log.Error(err)
		return retPeers, leechers, seeders, err
	}

	log.Debugf("peers num %d\n", len(t.Peers))
	for peer, info := range t.Peers {
		log.Debugf("peer %v\n", info)
		if info.Timestamp.After(now.Add(expireDuration)) {
			log.Debugf("before\n")
			if peersNotExpireCount < int(numWant) || numWant == DEFAULT_NUMWANT_TO_GET_ALL_PEERS {
				retPeers = append(retPeers, info)
			}
			peersNotExpire[peer] = info
			peersNotExpireCount += 1
		}
		peersCount += 1
	}

	log.Debugf("retPeers \n", retPeers)
	// peers update strategy, needs more design
	if peersCount >= MAX_PEERS_LENGTH {
		err = this.PutTorrent(fileHash, &Torrent{Leechers: t.Leechers, Seeders: t.Seeders, Peers: peersNotExpire})
		if err != nil {
			return retPeers, leechers, seeders, err
		}
	}

	return retPeers, leechers, seeders, nil
}

func (this *TorrentDB) GetTorrentPeerByFileHashAndNodeAddr(fileHash []byte, nodeAddr string) (peer *PeerInfo, err error) {
	t, err := this.GetTorrent(fileHash)
	if err != nil {
		return peer, err
	}
	if t == nil {
		return peer, errors.New("torrent not found, got nil")
	}

	peer, ok := t.Peers[nodeAddr]
	if !ok {
		return peer, errors.New("peer not found")
	}
	return peer, nil
}
