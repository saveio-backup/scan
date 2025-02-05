package tracker

import (
	"fmt"
	"net"
	"testing"

	chainComm "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

func TestUnsupportedTrackerScheme(t *testing.T) {
	t.Parallel()
	ret, err := Announce{TrackerUrl: "udp://public.popcorn-tracker.org:6969/announce"}.Do()
	fmt.Printf("resp:%v, err:%s\n", ret, err)
	// require.Equal(t, ErrBadScheme, err)
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
