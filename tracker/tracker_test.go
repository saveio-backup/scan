package tracker

import (
	"fmt"
	"net"
	"testing"

	"github.com/saveio/scan/storage"
	chainComm "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

func TestUnsupportedTrackerScheme(t *testing.T) {
	t.Parallel()
	ret, err := Announce{TrackerUrl: "udp://public.popcorn-tracker.org:6969/announce"}.Do()
	fmt.Printf("resp:%v, err:%s\n", ret, err)
	// require.Equal(t, ErrBadScheme, err)
}

func TestTorrentComplete(t *testing.T) {
	// args: infoHash, trackerUrl
	infoHash := storage.MetaInfoHash{}
	// copy(infoHash[:], []uint8{0xa3, 0x56, 0x41, 0x43, 0x74, 0x23, 0xe6, 0x26, 0xd9, 0x38, 0x25, 0x4a, 0x6b, 0x80, 0x49, 0x10, 0xa6, 0x67, 0xa, 0xc1})
	ids := "QmaRDZPe3QdnvCaUPUafk3EUMkWfsc4mtTosTDQQ9m4Ddc"
	idsBytes := []byte(ids)
	copy(infoHash[:], idsBytes)
	url := "udp://localhost:6369/announce"
	ip := net.IP{0x0, 0x1, 0x2, 0x3}
	port := uint16(6302)
	CompleteTorrent(infoHash, url, ip, port)
}

func TestGetTorrentPeers(t *testing.T) {
	infoHash := storage.MetaInfoHash{}
	// copy(infoHash[:], []uint8{0xa3, 0x56, 0x41, 0x43, 0x74, 0x23, 0xe6, 0x26, 0xd9, 0x38, 0x25, 0x4a, 0x6b, 0x80, 0x49, 0x10, 0xa6, 0x67, 0xa, 0xc1})
	// fmt.Println(string(infoHash[:]))
	// ids := "zb2rhmFsUmnSMrZodXs9vjjZePJPdxjVjXzbNRQNXpahe4"
	ids := "QmTrKW3x3Wmin1iAxdLQ8VTdCLY6uk2MLuFVYu7JHJSRdz"
	idsBytes := []byte(ids)
	copy(infoHash[:], idsBytes)
	// ids := "zb2rhmiu2V1kTDk5SRRo2F7b5WAivNDzQeDq7Qm3RNVndh5Gz"
	// id, err := cid.Parse(ids)
	// fmt.Printf("%v, %d, %+v\n", id.Bytes(), len(id.Bytes()), err)

	// idStr, err := cid.Parse(id.Bytes())
	// fmt.Printf("%s, %+v\n", idStr, err)

	// url := "udp://localhost:6369/announce"
	var trackerUrlOnline = "udp://40.73.102.177:6369/announce"
	peers := GetTorrentPeers(infoHash, trackerUrlOnline, -1, 1)
	fmt.Printf("peers:%v\n", peers)
}

func TestRegNodeType(t *testing.T) {
	walletAddr := chainComm.Address{0xa3, 0x56, 0x41, 0x43, 0x74, 0x23, 0xe6, 0x26, 0xd9, 0x38, 0x25, 0x4a, 0x6b, 0x80, 0x49, 0x10, 0xa6, 0x67, 0xa, 0xc1}
	netIp := net.ParseIP("127.0.0.1")
	url := "udp://localhost:6369/announce"

	err := RegNodeType(url, walletAddr, netIp, 3380, NODE_TYPE_DDNS)
	if err != nil {
		log.Error("TestRegNodeType error: ", err.Error())
	}
}

func TestGetNodesByType(t *testing.T) {
	url := "udp://localhost:6369/announce"
	nodesInfo, err := GetNodesByType(url, NODE_TYPE_DDNS)
	if err != nil {
		log.Error("TestRegNodeType error: ", err.Error())
	}
	fmt.Printf("nodesInfo: %v", nodesInfo)
}
