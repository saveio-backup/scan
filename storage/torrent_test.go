package storage

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/anacrolix/dht/krpc"
	"github.com/saveio/max/thirdparty/assert"
)

func TestPeerInfo_Serialize(t *testing.T) {
	id := PeerID{}
	rand.Read(id[:])

	pi := &PeerInfo{
		ID:       id,
		Complete: true,
		IP:       [4]byte{0x7F, 0x0, 0x0, 0x1},
		Port:     3380,
		NodeAddr: krpc.NodeAddr{
			IP:   []byte{0x7F, 0x0, 0x0, 0x1},
			Port: 3380,
		},
		Timestamp: time.Now(),
	}
	fmt.Printf("pi: %v\n", pi)

	var buffer bytes.Buffer
	err := pi.Serialize(&buffer)
	assert.Nil(err, t)
	fmt.Printf("buf: %v\n", buffer.Bytes())

	bin := buffer.Bytes()
	reader := bytes.NewReader(bin)
	dpi := &PeerInfo{}
	dpi.DeSerialize(reader)
	fmt.Printf("dpi: %v\n", dpi)
}
