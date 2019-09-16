package tracker

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"github.com/saveio/scan/storage"
	"github.com/saveio/themis/crypto/keypair"

	"github.com/saveio/max/thirdparty/assert"
	chainsdk "github.com/saveio/themis-go-sdk/utils"
	"github.com/saveio/themis-go-sdk/wallet"
	"github.com/saveio/themis/account"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

var wallet1Addr = "AYMnqA65pJFKAbbpD8hi5gdNDBmeFBy5hS"
var wallet2Addr = "AWaE84wqVf1yffjaR6VJ4NptLdqBAm8G9c"
var wallet3Addr = "AXUhmdzcAJwaFW91q6UYuPGGJY3fimoTAj"
var trackerUrl = "udp://localhost:6369/announce"
var trackerUrlOnline = "udp://52.187.109.234:6369/announce"
var walletfile = "../wallet.dat"
var walletpwd = "pwd"

func TestSignatureVerify(t *testing.T) {
	acc, err := GetAccount(walletfile, walletpwd)
	assert.Nil(err, t)
	fmt.Println(acc.Address.ToBase58())

	ip := net.ParseIP("192.168.1.1").To4()
	var ipAddr [4]byte
	copy(ipAddr[:4], ip[:4])
	fmt.Printf("%+v\n", len(ip))
	fmt.Printf("%+v\n", ipAddr)

	pkLength := make([]byte, 4)
	binary.BigEndian.PutUint16(pkLength, uint16(10000))
	fmt.Println(pkLength, binary.BigEndian.Uint16(pkLength))

	apiParams := ActionEndpointRegParams{
		Wallet: acc.Address,
		IP:     net.ParseIP("192.168.1.1").To4(),
		Port:   uint16(8888),
	}
	fmt.Println("pubKey", acc.PublicKey)

	rawData, err := json.Marshal(apiParams)
	fmt.Println("rawData", rawData, err)
	assert.Nil(err, t)

	sigData, err := chainsdk.Sign(acc, rawData)
	fmt.Println("sigData", sigData, err)

	pks := keypair.SerializePublicKey(acc.PublicKey)
	pk, err := keypair.DeserializePublicKey(pks)
	fmt.Println("pubKey Binarys", pk, err)

	err = chainsdk.Verify(pk, rawData, sigData)
	fmt.Println(err)
	assert.Nil(err, t)
}

func TestTorrentComplete(t *testing.T) {

	acc, err := GetAccount(walletfile, walletpwd)
	assert.Nil(err, t)
	fmt.Println(acc.Address.ToBase58())

	infoHash := storage.MetaInfoHash{}
	ids := "QmaRDZPe3QdnvCaUPUafk3EUMkWfsc4mtTosTDQQ9m4aaa"
	// ids := "zb2rhmiu2V1kTDk5SRRo2F7b5WAivNDzQeDq7Qm3RNVndh5Gz"
	// ids := "zb2rhmFsUmnSMrZodXs9vjjZePJPdxjVjXzbNRQNXpahe4"
	copy(infoHash[:], []byte(ids))
	apiParams := ActionTorrentCompleteParams{
		InfoHash: infoHash,
		IP:       net.ParseIP("192.168.1.1"),
		Port:     uint16(8888),
	}
	err = CompleteTorrent(trackerUrl, apiParams, acc.PublicKey, func(raw []byte) ([]byte, error) {
		return chainsdk.Sign(acc, raw)
	})
	assert.Nil(err, t)
}

func TestGetTorrentPeers(t *testing.T) {

	acc, err := GetAccount(walletfile, walletpwd)
	assert.Nil(err, t)
	fmt.Println(acc.Address.ToBase58())

	infoHash := storage.MetaInfoHash{}
	ids := "QmaRDZPe3QdnvCaUPUafk3EUMkWfsc4mtTosTDQQ9m4aaa"
	// ids := "zb2rhmiu2V1kTDk5SRRo2F7b5WAivNDzQeDq7Qm3RNVndh5Gz"
	// ids := "zb2rhmFsUmnSMrZodXs9vjjZePJPdxjVjXzbNRQNXpahe4"
	copy(infoHash[:], []byte(ids))
	apiParams := ActionGetTorrentPeersParams{
		InfoHash: infoHash,
		NumWant:  -1,
		Left:     1,
	}
	peers, err := GetTorrentPeers(trackerUrl, apiParams, acc.PublicKey, func(raw []byte) ([]byte, error) {
		return chainsdk.Sign(acc, raw)
	})
	assert.Nil(err, t)
	fmt.Printf("peers:%v\n", peers)
}

func TestRegEndPointForClient(t *testing.T) {
	acc, err := GetAccount(walletfile, walletpwd)
	assert.Nil(err, t)
	fmt.Println(acc.Address.ToBase58())

	apiParams := ActionEndpointRegParams{
		Wallet: acc.Address,
		IP:     net.ParseIP("192.168.1.1"),
		Port:   uint16(8887),
	}

	err = RegEndPoint(trackerUrlOnline, apiParams, acc.PublicKey, func(raw []byte) ([]byte, error) {
		return chainsdk.Sign(acc, raw)
	})
	assert.Nil(err, t)
}

func TestReqEndPointForClient(t *testing.T) {
	acc, err := GetAccount(walletfile, walletpwd)
	assert.Nil(err, t)
	fmt.Println(acc.Address.ToBase58())

	wb, err := common.AddressFromBase58(acc.Address.ToBase58())
	assert.Nil(err, t)
	hostAddr, err := ReqEndPoint(trackerUrlOnline, wb)
	assert.Nil(err, t)
	fmt.Printf("hostAddr:%s\n", hostAddr)
}

func GetAccount(w, wp string) (*account.Account, error) {
	wal, err := wallet.OpenWallet(w)
	if err != nil {
		log.Errorf("open wallet err:%s\n", err)
		return nil, nil
	}
	acc, err := wal.GetDefaultAccount([]byte(wp))
	if err != nil {
		log.Errorf("get default acc err:%s\n", err)
		return nil, nil
	}
	return acc, nil
}
