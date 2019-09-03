package storage

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"
	"time"

	"github.com/anacrolix/dht/krpc"
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
	PORT_LENGTH                      = 2
	TIMESTAMP_BINARY_LENGTH          = 8
	CHECKSUM_LEN                     = 4
	PAYLOAD_LEN                      = 39
	BUFFER_LEN                       = 50
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

func (this *PeerInfo) Serialize() []byte {
	var completeBuf [1]byte
	completeBuf[0] = byte(0)
	if this.Complete == true {
		completeBuf[0] = byte(1)
	}

	var portBuf [PORT_LENGTH]byte
	binary.LittleEndian.PutUint16(portBuf[:], this.Port)

	timestamp := make([]byte, TIMESTAMP_BINARY_LENGTH)
	var timestampBuf [TIMESTAMP_BINARY_LENGTH]byte
	u := uint64(this.Timestamp.Unix())
	binary.BigEndian.PutUint64(timestamp, u)
	copy(timestampBuf[:], timestamp[:TIMESTAMP_BINARY_LENGTH])

	var result []byte
	result = append(result, this.ID[:PEERID_LENGTH]...)
	result = append(result, completeBuf[0])
	result = append(result, this.IP[:IP_LENGTH]...)
	result = append(result, portBuf[:PORT_LENGTH]...)
	result = append(result, timestamp[:]...)

	checkSum := crc32.ChecksumIEEE(result)
	checkSumBuf := make([]byte, CHECKSUM_LEN)
	binary.BigEndian.PutUint32(checkSumBuf, checkSum)

	result = append(result, checkSumBuf...)

	prefixResult := make([]byte, BUFFER_LEN)
	copy(prefixResult, result)
	return prefixResult
}

func (this *PeerInfo) Deserialize(base64Buf []byte) error {
	if len(base64Buf) != BUFFER_LEN {
		return errors.New("buffer length illegal")
	}
	buf := make([]byte, BUFFER_LEN)
	copy(buf[:], base64Buf[:])

	payload := buf[:PAYLOAD_LEN-CHECKSUM_LEN]
	checkSum := crc32.ChecksumIEEE(payload)
	check := binary.BigEndian.Uint32(buf[PAYLOAD_LEN-CHECKSUM_LEN:])
	if checkSum != check {
		return errors.New("checkSum failed")
	}

	copy(this.ID[:], buf[:PEERID_LENGTH])
	completeBuf := buf[PEERID_LENGTH]
	this.Complete = false
	if completeBuf == byte(1) {
		this.Complete = true
	}

	completeEnd := PEERID_LENGTH + 1
	ipEnd := completeEnd + IP_LENGTH
	portEnd := ipEnd + PORT_LENGTH
	timeEnd := portEnd + TIMESTAMP_BINARY_LENGTH
	copy(this.IP[:], buf[completeEnd:ipEnd])

	var portBuf [PORT_LENGTH]byte
	copy(portBuf[:], buf[ipEnd:portEnd])
	this.Port = binary.LittleEndian.Uint16(portBuf[:])
	i := int64(binary.BigEndian.Uint64(buf[portEnd:timeEnd]))
	this.Timestamp = time.Unix(i, 0)

	this.NodeAddr = krpc.NodeAddr{
		IP:   this.IP[:],
		Port: int(this.Port),
	}
	return nil
}

func (this *PeerInfo) String() string {
	buf := this.Serialize()
	return string(buf)
}

func (this *PeerInfo) ParseFromString(base64Str string) {
	this.Deserialize([]byte(base64Str))
}

func (this *PeerInfo) Print() {
	log.Infof("ID: %v", this.ID)
	log.Infof("Complete: %v", this.Complete)
	log.Infof("NodeAddr: %s", this.NodeAddr)
	log.Infof("Timestamp: %s", this.Timestamp)
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

	log.Debugf("retPeers: %v\n", retPeers)
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
