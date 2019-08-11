package service

import (
	"container/list"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/dsp-go-sdk/channel"
	"github.com/saveio/dsp-go-sdk/store"
	chActorClient "github.com/saveio/pylons/actor/client"
	ch_actor "github.com/saveio/pylons/actor/server"
	chanCom "github.com/saveio/pylons/common"
	"github.com/saveio/scan/common/config"
	"github.com/saveio/scan/storage"
	chainCom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

func NewScanChannel(scan *Node, p2pActor *actor.PID) (*channel.Channel, error) {
	var dbstore *store.LevelDBStore
	if len(scan.Config.DBPath) > 0 {
		var err error
		dbstore, err = store.NewLevelDBStore(scan.Config.DBPath)
		if err != nil {
			log.Errorf("init db err %s", err)
			return nil, nil
		}
	}

	if len(scan.Config.ChannelListenAddr) > 0 && scan.Account != nil {
		var err error
		getHostCallBack := func(addr chainCom.Address) (string, error) {
			return GetExternalIP(addr.ToBase58())
		}

		ch, err := channel.NewChannelService(scan.Config, scan.Chain, getHostCallBack)
		if err != nil {
			log.Errorf("init channel err %s", err)
			return nil, err
		}
		chActorClient.SetP2pPid(p2pActor)
		if dbstore != nil {
			channelDB := store.NewChannelDB(dbstore)
			ch.SetChannelDB(channelDB)
		}
		return ch, nil
	}
	return nil, errors.New("not start")
}

func (this *Node) StartChannelService() error {
	if this.Channel == nil {
		return errors.New("channel is nil")
	}
	this.SetupPartnerHost(this.Channel.GetAllPartners())
	err := this.Channel.StartService()
	if err != nil {
		return err
	}
	time.Sleep(time.Second)
	this.Channel.OverridePartners()
	return nil
}

// SetupPartnerHost. setup host addr for partners
func (this *Node) SetupPartnerHost(partners []string) {
	for _, addr := range partners {
		host, err := GetExternalIP(addr)
		if err != nil && host != "" {
			log.Infof("SetHostAddr partners, addr: %s host: %s", addr, host)
			this.Channel.SetHostAddr(addr, host)
		}
	}
}

// GetExternalIP. get external ip of wallet from dns nodes
// func GetExternalIP(walletAddr string) (string, error) {
// 	w, _ := common.WHPTobyte(walletAddr, "")
// 	hpBytes, err := storage.TDB.Get(w)
// 	if err != nil {
// 		return "", err
// 	} else {
// 		var nodeAddr krpc.NodeAddr
// 		log.Infof("Channel.GetExternalIP wallAddr: %v, hpBytes: %v", walletAddr, hpBytes)
// 		nodeAddr.UnmarshalBinary(hpBytes)
// 		log.Infof("Channel.GetExternalIP nodeAddr: %s:%d", nodeAddr.IP, nodeAddr.Port)
// 		return fmt.Sprintf("%s://%s:%d", config.Parameters.Base.ChannelProtocol, nodeAddr.IP, nodeAddr.Port), nil
// 	}
// }

func GetExternalIP(walletAddr string) (string, error) {
	log.Debugf("channel get hostinfo call %s", walletAddr)
	nodeAddr, err := storage.EDB.GetEndpoint(walletAddr)
	log.Debugf("%v, err: %v", nodeAddr, err)
	if err != nil {
		return "", err
	}
	log.Debugf("%s://%v", config.Parameters.Base.ChannelProtocol, nodeAddr.NodeAddr)
	return fmt.Sprintf("%s://%v", config.Parameters.Base.ChannelProtocol, nodeAddr.NodeAddr), nil
}

func PutExternalIP(wallAddr, hostAddr string) error {
	index := strings.Index(hostAddr, "://")
	hostPort := hostAddr
	if index != -1 {
		hostPort = hostAddr[index+3:]
	}
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return err
	}
	netIp := net.ParseIP(host).To4()
	if netIp == nil {
		netIp = net.ParseIP(host).To16()
	}
	netPort, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	err = storage.EDB.PutEndpoint(wallAddr, netIp, netPort)
	if err != nil {
		return err
	}
	return nil
}

//pylons api
func (this *Node) OpenChannel(partnerAddr string, amount uint64) (chanCom.ChannelID, error) {
	return this.Channel.OpenChannel(partnerAddr, amount)
}

// func (this *ChannelSvr) CloseChannel(partnerAddr string) error {
// 	return this.Channel.ChannelClose(partnerAddr)
// }

func (this *Node) QuerySpecialChannelDeposit(partnerAddr string) (uint64, error) {
	return this.Channel.GetTotalDepositBalance(partnerAddr)
}

func (this *Node) DepositToChannel(partnerAddr string, totalDeposit uint64) error {
	return this.Channel.SetDeposit(partnerAddr, totalDeposit)
}

func (this *Node) GetAllChannels() (*ch_actor.ChannelsInfoResp, error) {
	return this.Channel.AllChannels()
}

func (this *Node) Transfer(paymentId int32, amount uint64, to string) error {
	return this.Channel.DirectTransfer(paymentId, amount, to)
}

func (this *Node) GetChannelListByOwnerAddress(addr string, tokenAddr string) *list.List {
	//[TODO] call dsp-go-sdk function to return channel list
	//[NOTE] addr and token Addr should NOT be needed. addr mean PaymentNetworkID
	//tokenAddr mean TokenAddress. Need comfirm the behavior when integrate dsp-go-sdk with pylons
	return list.New()
}

func (this *Node) ChannelWithdraw(partnerAddr string, amount uint64) error {
	success, err := this.Channel.Withdraw(partnerAddr, amount)
	if err != nil {
		return err
	}
	if !success {
		return errors.New("withdraw failed")
	}
	return nil
}

func (this *Node) QueryHostInfo(partnerAddr string) (string, error) {
	return this.Channel.GetHostAddr(partnerAddr)
}

var startChannelHeight uint32

type FilterBlockProgress struct {
	Progress float32
	Start    uint32
	End      uint32
	Now      uint32
}

func (this *Node) GetFilterBlockProgress() (*FilterBlockProgress, error) {
	progress := &FilterBlockProgress{}
	if this.Channel == nil {
		return progress, nil
	}
	endChannelHeight, err := this.Chain.GetCurrentBlockHeight()
	if err != nil {
		log.Debugf("get channel err %s", err)
		return progress, err
	}
	if endChannelHeight == 0 {
		return progress, nil
	}
	progress.Start = startChannelHeight
	progress.End = endChannelHeight
	now := this.Channel.GetCurrentFilterBlockHeight()
	progress.Now = now
	log.Debugf("endChannelHeight %d, start %d", endChannelHeight, startChannelHeight)
	if endChannelHeight <= startChannelHeight {
		progress.Progress = 1.0
		return progress, nil
	}
	rangeHeight := endChannelHeight - startChannelHeight
	if now >= rangeHeight+startChannelHeight {
		progress.Progress = 1.0
		return progress, nil
	}
	p := float32(now-startChannelHeight) / float32(rangeHeight)
	progress.Progress = p
	log.Debugf("GetFilterBlockProgress start %d, now %d, end %d, progress %v", startChannelHeight, now, endChannelHeight, progress)
	return progress, nil
}
