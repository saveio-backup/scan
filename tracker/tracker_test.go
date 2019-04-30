package tracker

import (
	"fmt"
	"net"
	"testing"

	"github.com/saveio/scan/common"
)

func TestUnsupportedTrackerScheme(t *testing.T) {
	t.Parallel()
	ret, err := Announce{TrackerUrl: "udp://public.popcorn-tracker.org:6969/announce"}.Do()
	fmt.Printf("resp:%v, err:%s\n", ret, err)
	// require.Equal(t, ErrBadScheme, err)
}

func TestTorrentComplete(t *testing.T) {
	// args: infoHash, trackerUrl
	infoHash := common.MetaInfoHash{}
	copy(infoHash[:], []uint8{0xa3, 0x56, 0x41, 0x43, 0x74, 0x23, 0xe6, 0x26, 0xd9, 0x38, 0x25, 0x4a, 0x6b, 0x80, 0x49, 0x10, 0xa6, 0x67, 0xa, 0xc1})
	url := "udp://localhost:6369/announce"
	ip := net.IP{0x0, 0x1, 0x2, 0x3}
	port := uint16(6370)
	CompleteTorrent(infoHash, url, ip, port)
}

func TestGetTorrentPeers(t *testing.T) {
	infoHash := common.MetaInfoHash{}
	copy(infoHash[:], []uint8{0xa3, 0x56, 0x41, 0x43, 0x74, 0x23, 0xe6, 0x26, 0xd9, 0x38, 0x25, 0x4a, 0x6b, 0x80, 0x49, 0x10, 0xa6, 0x67, 0xa, 0xc1})
	url := "udp://localhost:6369/announce"
	peers := GetTorrentPeers(infoHash, url, -1, 1)
	fmt.Printf("peers:%v\n", peers)
}
