package test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/saveio/carrier/crypto"
	"github.com/saveio/carrier/crypto/ed25519"
	"github.com/saveio/max/thirdparty/assert"
	pm "github.com/saveio/scan/p2p/actor/messages"
	tkActClient "github.com/saveio/scan/p2p/actor/tracker/client"
	tkActServer "github.com/saveio/scan/p2p/actor/tracker/server"
	tk_net "github.com/saveio/scan/p2p/networks/tracker"
	"github.com/saveio/scan/service/tk"
	"github.com/saveio/scan/storage"
	chainsdk "github.com/saveio/themis-go-sdk/utils"
	"github.com/saveio/themis-go-sdk/wallet"
	"github.com/saveio/themis/account"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/crypto/keypair"
)

var natProxyServerAddr = "tcp://40.73.100.114:6007"
var tkListenAddr = "tcp://127.0.0.1:10887"
var targetDnsAddr = "tcp://40.73.100.114:30668"
var walletFile = "./wallet.dat"
var walletPwd = "pwd"

var ids = "QmaRDZPe3QdnvCaUPUafk3EUMkWfsc4mtTosTDQQ9m4aaa"

// var ids = "zb2rhmiu2V1kTDk5SRRo2F7b5WAivNDzQeDq7Qm3RNVndh5Gz"
// var ids = "zb2rhmFsUmnSMrZodXs9vjjZePJPdxjVjXzbNRQNXpahe4"

func TestAnnounceRequestCompleteTorrent(t *testing.T) {
	tkActSrv, _ := InitializeService()
	annResp, err := tkActSrv.AnnounceRequestCompleteTorrent(&pm.CompleteTorrentReq{
		InfoHash: []byte(ids),
		Ip:       net.ParseIP("192.168.1.1"),
		Port:     uint64(1111),
	}, targetDnsAddr)
	log.Debugf("announce response: %v, err %v\n", annResp, err)
	assert.Nil(err, nil)

	// WaitToExit()
}

func TestAnnounceRequestTorrentPeers(t *testing.T) {
	tkActSrv, _ := InitializeService()
	annResp, err := tkActSrv.AnnounceRequestTorrentPeers(&pm.GetTorrentPeersReq{
		InfoHash: []byte(ids),
		NumWant:  2,
	}, targetDnsAddr)
	log.Debugf("announce response: %v, err %v\n", annResp, err)
	assert.Nil(err, nil)

	// WaitToExit()
}

func TestAnnounceRequestEndpointRegistry(t *testing.T) {
	tkActSrv, acc := InitializeService()
	annResp, err := tkActSrv.AnnounceRequestEndpointRegistry(&pm.EndpointRegistryReq{
		Wallet: acc.Address[:],
		Ip:     net.ParseIP("192.168.1.1"),
		Port:   uint64(1111),
	}, targetDnsAddr)
	log.Debugf("announce response: %v, err %v\n", annResp, err)
	assert.Nil(err, nil)

	// WaitToExit()
}

func TestAnnounceRequestGetEndpointAddr(t *testing.T) {
	tkActSrv, acc := InitializeService()
	annResp, err := tkActSrv.AnnounceRequestGetEndpointAddr(&pm.QueryEndpointReq{
		Wallet: acc.Address[:],
	}, targetDnsAddr)
	log.Debugf("announce response: %v, err %v\n", annResp, err)
	assert.Nil(err, nil)

	// WaitToExit()
}

func InitializeService() (*tkActServer.TrackerActorServer, *account.Account) {
	log.InitLog(1, log.Stdout)
	acc, err := GetAccount(walletFile, walletPwd)
	assert.Nil(err, nil)
	fmt.Println(acc.Address.ToBase58())

	tkSrv := tk.NewTrackerService(nil, acc.PublicKey, func(raw []byte) ([]byte, error) {
		return chainsdk.Sign(acc, raw)
	})

	tkActSrv, err := StartTkActServer(tkSrv, acc)
	assert.Nil(err, nil)

	tkActClient.SetTrackerServerPid(tkActSrv.GetLocalPID())
	go tkSrv.Start(targetDnsAddr)

	// cilent does't needs to do db setup
	tdb, err := storage.NewLevelDBStore(filepath.Join("../DB", acc.Address.ToBase58(), "tracker"))
	if err != nil {
		log.Fatal(err)
	}
	storage.TDB = storage.NewTorrentDB(tdb)

	edb, err := storage.NewLevelDBStore(filepath.Join("../DB", acc.Address.ToBase58(), "endpoint"))
	if err != nil {
		log.Fatal(err)
	}
	storage.EDB = storage.NewEndpointDB(edb)
	return tkActSrv, acc
}

func StartTkActServer(tkSrc *tk.TrackerService, acc *account.Account) (*tkActServer.TrackerActorServer, error) {
	tkActServer, err := tkActServer.NewTrackerActor(tkSrc)
	fmt.Println(tkActServer, err)
	if err != nil {
		return nil, err
	}

	dPub := keypair.SerializePublicKey(acc.PubKey())
	tkPub, tkPri, err := ed25519.GenerateKey(&accountReader{
		PublicKey: append(dPub, []byte("tk")...),
	})
	tkNetworkKey := &crypto.KeyPair{
		PublicKey:  tkPub,
		PrivateKey: tkPri,
	}
	tkNet := tk_net.NewP2P()
	tkNet.SetNetworkKey(tkNetworkKey)
	tkNet.SetProxyServer(natProxyServerAddr)
	tkNet.SetPID(tkActServer.GetLocalPID())
	log.Infof("goto start tk network %s", tkListenAddr)
	tk_net.TkP2p = tkNet
	tkActServer.SetNetwork(tkNet)

	err = tkNet.Start(tkListenAddr)
	if err != nil {
		return nil, err
	}
	log.Infof("tk network started, public ip %s", tkNet.PublicAddr())
	return tkActServer, nil
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

type accountReader struct {
	PublicKey []byte
}

func (this accountReader) Read(buf []byte) (int, error) {
	bufs := make([]byte, 0)
	hash := sha256.Sum256(this.PublicKey)
	bufs = append(bufs, hash[:]...)
	log.Debugf("bufs :%s", hex.EncodeToString(bufs))
	for i, _ := range buf {
		if i < len(bufs) {
			buf[i] = bufs[i]
			continue
		}
		buf[i] = 0
	}
	return len(buf), nil
}

func WaitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("seeds received exit signal:%v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}
