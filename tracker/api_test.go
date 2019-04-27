/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-07
 */
package tracker

import (
	"fmt"
	"github.com/anacrolix/dht/krpc"
	"github.com/oniio/oniChain/common"
	"github.com/oniio/oniFS/thirdparty/assert"
	"net"
	"testing"
	//"github.com/oniio/oniChain/common/log"
)

var wallet1Addr = "AYMnqA65pJFKAbbpD8hi5gdNDBmeFBy5hS"
var wallet2Addr = "AWaE84wqVf1yffjaR6VJ4NptLdqBAm8G9c"
var trackerUrl = "udp://localhost:6369/announce"

func TestRegEndPoint(t *testing.T) {
	url := trackerUrl
	//ip:=net.ParseIP("192.168.1.1")
	ip := net.IP{0x11, 0x1, 0x2, 0x3}
	fmt.Printf("ip:%s\n", ip)
	fmt.Println(ip)
	port := uint16(8840)
	wb, _ := common.AddressFromBase58(wallet1Addr)
	err := RegEndPoint(url, wb, ip, port)
	assert.Nil(err, t)
}

func TestUnRegEndPoint(t *testing.T) {
	url := trackerUrl
	wb, _ := common.AddressFromBase58(wallet1Addr)
	err := UnRegEndPoint(url, wb)
	assert.Nil(err, t)
}

func TestReqEndPoint(t *testing.T) {
	url := trackerUrl
	wb, _ := common.AddressFromBase58(wallet1Addr)
	nb, err := ReqEndPoint(url, wb)
	assert.Nil(err, t)
	var nodeAddr krpc.NodeAddr
	nodeAddr.UnmarshalBinary(nb)
	fmt.Printf("ip:%s\n", nodeAddr.IP.String())
	fmt.Printf("port:%d\n", nodeAddr.Port)
}

func TestUpdateEndPoint(t *testing.T) {
	url := trackerUrl
	//ip:=net.ParseIP("192.168.1.1")
	ip := net.IP{0x11, 0x11, 0x12, 0x3}
	fmt.Printf("ip:%s\n", ip)
	fmt.Println(ip)
	port := uint16(8840)
	wb, _ := common.AddressFromBase58(wallet2Addr)
	err := UpdateEndPoint(url, wb, ip, port)
	assert.Nil(err, t)
}
