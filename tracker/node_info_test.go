package tracker

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/saveio/max/thirdparty/assert"
	chainCom "github.com/saveio/themis/common"
)

func TestNodesInfoSt_Serialize(t *testing.T) {
	var err error
	var nodeNetAddr1 [NODENETADDR]byte
	var nodeNetAddr2 [NODENETADDR]byte
	copy(nodeNetAddr1[:], []byte("127.0.0.1:3380"))
	copy(nodeNetAddr2[:], []byte("127.0.0.1:3381"))
	walletAddr1 := chainCom.Address{0xa3, 0x56, 0x41, 0x43, 0x74, 0x23, 0xe6, 0x26, 0xd9, 0x38, 0x25, 0x4a, 0x6b, 0x80, 0x49, 0x10, 0xa6, 0x67, 0xa, 0xc1}
	walletAddr2 := chainCom.Address{0xa3, 0x56, 0x41, 0x43, 0x74, 0x23, 0xe6, 0x26, 0xd9, 0x38, 0x25, 0x4a, 0x6b, 0x80, 0x49, 0x10, 0xa6, 0x67, 0xa, 0xc2}

	fmt.Printf("nodeNetAddr1: %v\n", nodeNetAddr1)
	fmt.Printf("walletAddr1: %v\n", walletAddr1)

	st := &NodesInfoSt{
		NodeInfoLen: 2,
		NodeInfos: []NodeInfo{
			{
				NodeNetAddr: nodeNetAddr1,
				WalletAddr:  walletAddr1,
			},
			{
				NodeNetAddr: nodeNetAddr2,
				WalletAddr:  walletAddr2,
			},
		},
	}
	fmt.Printf("st: %v\n", st)

	var buffer bytes.Buffer
	err = st.Serialize(&buffer)
	assert.Nil(err, t)
	//fmt.Printf("Buf: %v", buffer.Bytes())

	bin := buffer.Bytes()
	reader := bytes.NewReader(bin)
	dst := &NodesInfoSt{}
	dst.DeSerialize(reader)
	fmt.Printf("dst: %v\n", dst)
}
