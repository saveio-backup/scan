/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-07
 */
package tracker

import (
	"fmt"
	"github.com/anacrolix/dht/krpc"
	"github.com/saveio/themis/common"
	"github.com/saveio/max/thirdparty/assert"
	"net"
	"testing"
	//"github.com/saveio/themis/common/log"
)

var walletAddr1 = "AYMnqA65pJFKAbbpD8hi5gdNDBmeFBy5hS"
var walletAddr2 = "AWaE84wqVf1yffjaR6VJ4NptLdqBAm8G9c"
var trackerUrl1 = "udp://localhost:6369/announce"

func TestRegEndPoint(t *testing.T) {
	url := trackerUrl1
	//ip:=net.ParseIP("192.168.1.1")
	ip := net.IP{0x11, 0x1, 0x2, 0x3}
	fmt.Printf("ip:%s\n", ip)
	fmt.Println(ip)
	port := uint16(8840)
	wb, _ := common.AddressFromBase58(walletAddr1)
	err := RegEndPoint(url, nil, nil, wb, ip, port)
	assert.Nil(err, t)
}

func TestUnRegEndPoint(t *testing.T) {
	url := trackerUrl1
	wb, _ := common.AddressFromBase58(walletAddr1)
	err := UnRegEndPoint(url, wb)
	assert.Nil(err, t)
}

func TestReqEndPointEx(t *testing.T) {
	url := trackerUrl1
	wb, _ := common.AddressFromBase58(walletAddr1)
	nb, err := ReqEndPoint(url, wb)
	assert.Nil(err, t)
	var nodeAddr krpc.NodeAddr
	nodeAddr.UnmarshalBinary(nb)
	fmt.Printf("ip:%s\n", nodeAddr.IP.String())
	fmt.Printf("port:%d\n", nodeAddr.Port)
}

func TestUpdateEndPointEx(t *testing.T) {
	url := trackerUrl1
	//ip:=net.ParseIP("192.168.1.1")
	ip := net.IP{0x11, 0x11, 0x12, 0x3}
	fmt.Printf("ip:%s\n", ip)
	fmt.Println(ip)
	port := uint16(8840)
	wb, _ := common.AddressFromBase58(walletAddr2)
	err := UpdateEndPoint(url, wb, ip, port)
	assert.Nil(err, t)
}
