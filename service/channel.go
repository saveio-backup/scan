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

	err := this.Channel.StartService()
	if err != nil {
		return err
	}
	time.Sleep(time.Second)
	this.Channel.OverridePartners()
	return nil
}

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

func (this *Node) OpenChannel(partnerAddr string, amount uint64) (chanCom.ChannelID, error) {
	return this.Channel.OpenChannel(partnerAddr, amount)
}

func (this *Node) CloseChannel(partnerAddr string) error {
	return this.Channel.ChannelClose(partnerAddr)
}

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

func (this *Node) channelExists(ci []*ch_actor.ChannelInfo, w string) bool {
	for _, ch := range ci {
		if ch.Address == w {
			return true
		}
	}
	return false
}
