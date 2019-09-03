package storage

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/anacrolix/dht/krpc"
)

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
	err := dpi.Deserialize(buffer)
	if err != nil {
		fmt.Println(err)
	}
	dpi.Print()
}
