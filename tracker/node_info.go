package tracker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/ontio/ontology/common/serialization"
	"github.com/saveio/scan/storage"
	chainCom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

const (
	NODENETADDR = 128
)

type NodeType int64

const (
	NODE_TYPE_DDNS    NodeType = 110
	NODE_TYPE_PORTER  NodeType = 111
	NODE_TYPE_STORAGE NodeType = 112
)

type NodeInfo struct {
	NodeNetAddr [NODENETADDR]byte
	WalletAddr  chainCom.Address
}

type NodesInfoSt struct {
	NodeInfoLen uint64
	NodeInfos   []NodeInfo
}

//buf := new(bytes.Buffer)
func (this *NodeInfo) Serialize(w io.Writer) error {
	var err error
	if err = serialization.WriteVarBytes(w, this.NodeNetAddr[:NODENETADDR]); err != nil {
		return fmt.Errorf("[NodeInfo] NodeNetAddr Serialize error: %s", err.Error())
	}

	if err = serialization.WriteVarBytes(w, this.WalletAddr[:chainCom.ADDR_LEN]); err != nil {
		return fmt.Errorf("[NodeInfo] WalletAddr Serialize error: %s", err.Error())
	}
	return nil
}

func (this *NodeInfo) DeSerialize(r io.Reader) error {
	var err error
	nodeNetAddr, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("[NodeInfo] NodeNetAddr DeSerialize error: %s", err.Error())
	}

	copy(this.NodeNetAddr[:], nodeNetAddr)
	walletAddr, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("[NodeInfo] walletAddr DeSerialize error: %s", err.Error())
	}
	this.WalletAddr, err = chainCom.AddressParseFromBytes(walletAddr)
	if err != nil {
		return fmt.Errorf("[NodeInfo] WalletAddr AddressParseFromBytes DeSerialize error: %s", err.Error())
	}
	return nil
}

//buf := new(bytes.Buffer)
func (this *NodesInfoSt) Serialize(w io.Writer) error {
	var err error
	if err = serialization.WriteUint64(w, this.NodeInfoLen); err != nil {
		return fmt.Errorf("[NodesInfoSt] NodeInfoLen Serialize error: %s", err.Error())
	}

	for i := uint64(0); i < this.NodeInfoLen; i++ {
		if err = this.NodeInfos[i].Serialize(w); err != nil {
			return err
		}
	}
	return err
}

func (this *NodesInfoSt) DeSerialize(r io.Reader) error {
	var err error
	if this.NodeInfoLen, err = serialization.ReadUint64(r); err != nil {
		return err
	}
	for i := uint64(0); i < this.NodeInfoLen; i++ {
		var nodeInfo NodeInfo
		if err := nodeInfo.DeSerialize(r); err != nil {
			return err
		}
		this.NodeInfos = append(this.NodeInfos, nodeInfo)
	}
	return err
}

func SaveNodeInfo(nodeType NodeType, nodeWallet chainCom.Address, nodeNetAddr string) error {

	var key []byte
	nodeTypeBuf := Int64ToBytes(nodeType)

	key = append(key, nodeTypeBuf[:]...)
	key = append(key, nodeWallet[:]...)

	log.Debugf("SaveNodeInfo WallerAddr %s, NodeAddr : %s", nodeWallet.ToBase58(), nodeNetAddr)
	return storage.EDB.Db.Put(key, []byte(nodeNetAddr))
}

func GetNodesInfo(nodeType NodeType) ([]byte, error) {
	var err error

	prefix := Int64ToBytes(nodeType)
	prefixLen := len(prefix)

	storeIterator := storage.EDB.Db.NewIterator(prefix[:])
	defer storeIterator.Release()

	var nodesInfos NodesInfoSt

	for storeIterator.Next() {
		key := storeIterator.Key()
		nodeWallerAddr, err := chainCom.AddressParseFromBytes(key[prefixLen:])
		if err != nil {
			log.Error("GetNodesInfo AddressParseFromBytes error: ", err.Error())
			continue
		}

		nodeInfo := NodeInfo{
			WalletAddr: nodeWallerAddr,
		}

		nodeNetAddr := storeIterator.Value()
		if len(nodeNetAddr) != 0 {
			copy(nodeInfo.NodeNetAddr[:], nodeNetAddr[:])
			nodesInfos.NodeInfos = append(nodesInfos.NodeInfos, nodeInfo)
			nodesInfos.NodeInfoLen++

			log.Debugf("GetNodesInfo WallerAddr %s, NodeAddr : %s", nodeWallerAddr.ToBase58(), nodeInfo.NodeNetAddr)
		} else {
			log.Error("GetNodesInfo len(nodeInfo.NodeNetAddr) == 0")
		}
	}
	if nodesInfos.NodeInfoLen == 0 {
		log.Debugf("GetNodesInfo error: no information")
	}

	buf := new(bytes.Buffer)
	if err = nodesInfos.Serialize(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func Int64ToBytes(i NodeType) [16]byte {
	var buf [16]byte
	copy(buf[:8], []byte("NodeType")[:8])
	binary.BigEndian.PutUint64(buf[8:], uint64(i))
	return buf
}
