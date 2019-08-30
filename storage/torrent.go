package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/anacrolix/dht/krpc"
	"github.com/ontio/ontology/common/serialization"
	"github.com/saveio/scan/common/config"
	"github.com/saveio/themis/common/log"
)

var TDB *TorrentDB

const (
	MAX_PEERS_LENGTH                 = 100
	DEFAULT_NUMWANT_TO_GET_ALL_PEERS = -1
	PEERID_LENGTH                    = 20
	METAINFOHASH_LENGTH              = 46
	IP_LENGTH                        = 4
)

type MetaInfoHash [METAINFOHASH_LENGTH]byte
type PeerID [PEERID_LENGTH]byte

type TorrentDB struct {
	db *LevelDBStore
}

type PeerInfo struct {
	ID        PeerID
	Complete  bool
	IP        [IP_LENGTH]byte
	Port      uint16
	NodeAddr  krpc.NodeAddr
	Timestamp time.Time
}

func (this *PeerInfo) Serialize(w io.Writer) error {
	var err error
	if err = serialization.WriteVarBytes(w, this.ID[:PEERID_LENGTH]); err != nil {
		return fmt.Errorf("[PeerInfo] ID Serialize error: %s", err.Error())
	}

	if err = serialization.WriteBool(w, this.Complete); err != nil {
		return fmt.Errorf("[PeerInfo] Complete Serialize error: %s", err.Error())
	}

	if err = serialization.WriteVarBytes(w, this.IP[:IP_LENGTH]); err != nil {
		return fmt.Errorf("[PeerInfo] IP Serialize error: %s", err.Error())
	}

	if err = serialization.WriteUint16(w, this.Port); err != nil {
		return fmt.Errorf("[PeerInfo] Port Serialize error: %s", err.Error())
	}

	naBinarys, err := this.NodeAddr.MarshalBinary()
	if err != nil {
		return fmt.Errorf("[PeerInfo] NodeAddr Serialize error: %s", err.Error())
	}
	if err = serialization.WriteVarBytes(w, naBinarys); err != nil {
		return fmt.Errorf("[PeerInfo] NodeAddrBinarys Serialize error: %s", err.Error())
	}

	tsBinarys, err := this.Timestamp.MarshalBinary()
	if err != nil {
		return fmt.Errorf("[PeerInfo] Timestamp Serialize error: %s", err.Error())
	}
	if err = serialization.WriteVarBytes(w, tsBinarys); err != nil {
		return fmt.Errorf("[PeerInfo] TimestampBynarys Serialize error: %s", err.Error())
	}
	return nil
}

func (this *PeerInfo) DeSerialize(r io.Reader) error {
	var err error
	ID, err := serialization.ReadVarBytes(r)
	if err != nil || ID == nil {
		return fmt.Errorf("[PeerInfo] ID DeSerialize error: %s", err.Error())
	}
	copy(this.ID[:], ID)

	this.Complete, err = serialization.ReadBool(r)
	if err != nil {
		return fmt.Errorf("[PeerInfo] Complete DeSerialize error: %s", err.Error())
	}

	IP, err := serialization.ReadVarBytes(r)
	if err != nil || IP == nil {
		return fmt.Errorf("[PeerInfo] IP DeSerialize error: %s", err.Error())
	}
	copy(this.IP[:], IP[:IP_LENGTH])

	this.Port, err = serialization.ReadUint16(r)
	if err != nil {
		return fmt.Errorf("[PeerInfo] ID Port error: %s", err.Error())
	}

	naBinarys, err := serialization.ReadVarBytes(r)
	if err != nil || naBinarys == nil {
		return fmt.Errorf("[PeerInfo] ID NodeAddrBinarys error: %s", err.Error())
	}
	err = this.NodeAddr.UnmarshalBinary(naBinarys)
	if err != nil {
		return fmt.Errorf("[PeerInfo] ID NodeAddr error: %s", err.Error())
	}

	taBinarys, err := serialization.ReadVarBytes(r)
	if err != nil || taBinarys == nil {
		return fmt.Errorf("[PeerInfo] ID TimestampBinary error: %s", err.Error())
	}
	err = this.Timestamp.UnmarshalBinary(taBinarys)
	if err != nil {
		return fmt.Errorf("[PeerInfo] ID Timestamp error: %s", err.Error())
	}
	return nil
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
