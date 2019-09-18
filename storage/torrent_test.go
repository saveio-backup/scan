package storage

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/anacrolix/dht/krpc"
	"github.com/saveio/max/thirdparty/assert"
)

var torrentDBPath string = "../testdata/db2"
var fileHash string = "zb2rhoYcu3YqAMWFA2WNj4qUZXqxUdRQqRQ5CvXGdabDanJpK"

func TestPeerInfo_Serialize(t *testing.T) {
	id := PeerID{0x7F, 0x0, 0x0, 0x1}
	rand.Read(id[:])

	pi := PeerInfo{
		ID:       [20]byte{210, 54, 81, 214, 82, 25, 40, 187, 12, 154, 16, 40, 119, 208, 68, 194, 55, 222, 113, 130},
		Complete: false,
		IP:       [4]byte{40, 73, 103, 72},
		Port:     30062,
		NodeAddr: krpc.NodeAddr{
			IP:   []byte{40, 73, 103, 72},
			Port: 30062,
		},
		Timestamp: time.Now(),
	}
	pi.Print()
	var buffer []byte
	buffer = pi.Serialize()
	fmt.Printf("buf: %v\n", buffer)
	dpi := PeerInfo{}
	// var dpi *PeerInfo
	err := dpi.Deserialize(buffer)
	if err != nil {
		fmt.Println(err)
	}
	dpi.Print()
}

func TestPutTorrent(t *testing.T) {
	db, err := NewLevelDBStore(torrentDBPath)
	if err != nil || db == nil {
		return
	}
	toDB := NewTorrentDB(db)
	fileHash := []byte("")
	torrent := &Torrent{
		Leechers: 0,
		Seeders:  0,
		Peers:    map[string]*PeerInfo{},
	}
	err = toDB.PutTorrent(fileHash, torrent)
	assert.Nil(err, t)
}

func TestGetTorrent(t *testing.T) {
	db, err := NewLevelDBStore(torrentDBPath)
	if err != nil || db == nil {
		return
	}
	toDB := NewTorrentDB(db)
	fileHash := []byte("")
	torrent, err := toDB.GetTorrent(fileHash)
	assert.Nil(err, t)
	fmt.Println("torrent: ", torrent)
}

func TestGetTorrentBinary(t *testing.T) {
	db, err := NewLevelDBStore(torrentDBPath)
	if err != nil || db == nil {
		return
	}
	toDB := NewTorrentDB(db)
	fileHash := []byte("")
	torrentBinary, err := toDB.GetTorrentBinary(fileHash)
	assert.Nil(err, t)
	fmt.Println("torrent binary: ", torrentBinary)
}

func TestDelTorrent(t *testing.T) {
	db, err := NewLevelDBStore(torrentDBPath)
	if err != nil || db == nil {
		return
	}
	toDB := NewTorrentDB(db)
	fileHash := []byte("")
	err = toDB.DelTorrent(fileHash)
	assert.Nil(err, t)
}

func TestAddTorrentPeer(t *testing.T) {
	db, err := NewLevelDBStore(torrentDBPath)
	if err != nil || db == nil {
		return
	}
	toDB := NewTorrentDB(db)
	fileHash := []byte("")
	err = toDB.AddTorrentPeer(fileHash, 1, "", &PeerInfo{
		ID:       PeerID{},
		Complete: true,
		IP:       [4]byte{40, 73, 103, 72},
		Port:     30062,
		NodeAddr: krpc.NodeAddr{
			IP:   []byte{40, 73, 103, 72},
			Port: 30062,
		},
		Timestamp: time.Now(),
	})
	assert.Nil(err, t)
}

func TestDelTorrentPeer(t *testing.T) {
	db, err := NewLevelDBStore(torrentDBPath)
	if err != nil || db == nil {
		return
	}
	toDB := NewTorrentDB(db)
	fileHash := []byte("")
	err = toDB.DelTorrentPeer(fileHash, &PeerInfo{
		ID:       PeerID{},
		Complete: true,
		IP:       [4]byte{40, 73, 103, 72},
		Port:     30062,
		NodeAddr: krpc.NodeAddr{
			IP:   []byte{40, 73, 103, 72},
			Port: 30062,
		},
		Timestamp: time.Now(),
	})
	assert.Nil(err, t)
}

func GetTorrentPeersByFileHash(t *testing.T) {
	db, err := NewLevelDBStore(torrentDBPath)
	if err != nil || db == nil {
		return
	}
	toDB := NewTorrentDB(db)
	fileHash := []byte("")
	peerInfo, leechers, seeders, err := toDB.GetTorrentPeersByFileHash(fileHash, -1)
	assert.Nil(err, t)
	fmt.Println("peerInfo: ", peerInfo)
	fmt.Println("leechers: ", leechers)
	fmt.Println("seeders: ", seeders)
}

func GetTorrentPeerByFileHashAndNodeAddr(t *testing.T) {
	db, err := NewLevelDBStore(torrentDBPath)
	if err != nil || db == nil {
		return
	}
	toDB := NewTorrentDB(db)
	fileHash := []byte("")
	peerInfo, err := toDB.GetTorrentPeerByFileHashAndNodeAddr(fileHash, "")
	assert.Nil(err, t)
	fmt.Println("peerInfo: ", peerInfo)
}
