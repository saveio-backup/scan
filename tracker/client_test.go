package tracker

import (
	"fmt"
	"net"
	"testing"

	"github.com/saveio/max/thirdparty/assert"
	"github.com/saveio/themis/common"
	//"github.com/saveio/themis/common/log"
)

var wallet1Addr = "AYMnqA65pJFKAbbpD8hi5gdNDBmeFBy5hS"
var wallet2Addr = "AWaE84wqVf1yffjaR6VJ4NptLdqBAm8G9c"
var trackerUrl = "udp://localhost:6369/announce"

func TestRegEndPointForClient(t *testing.T) {
	url := trackerUrl
	//ip:=net.ParseIP("192.168.1.1")
	ip := net.IP{0x11, 0x1, 0x2, 0x3}
	fmt.Printf("ip:%s\n", ip)
	fmt.Println(ip)
	port := uint16(8847)
	wb, _ := common.AddressFromBase58(wallet1Addr)
	fmt.Println(url, port, wb)
	err := RegEndPoint(url, nil, nil, wb, ip, port)
	assert.Nil(err, t)
}

func TestUnRegEndPointForClient(t *testing.T) {
	url := trackerUrl
	wb, _ := common.AddressFromBase58(wallet1Addr)
	err := UnRegEndPoint(url, wb)
	assert.Nil(err, t)
}

func TestReqEndPointForClient(t *testing.T) {
	url := trackerUrl
	wb, _ := common.AddressFromBase58(wallet1Addr)
	hostAddr, err := ReqEndPoint(url, wb)
	assert.Nil(err, t)
	fmt.Printf("hostAddr:%s\n", hostAddr)
}
