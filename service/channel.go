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
	"github.com/saveio/dsp-go-sdk/core/channel"
	"github.com/saveio/dsp-go-sdk/store"
	chActorClient "github.com/saveio/pylons/actor/client"
	ch_actor "github.com/saveio/pylons/actor/server"
	chanCom "github.com/saveio/pylons/common"
	"github.com/saveio/scan/common/config"
	"github.com/saveio/scan/storage"
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
		ch, err := channel.NewChannelService(
			scan.Chain,
			channel.ClientType(scan.Config.ChannelClientType),
			channel.RevealTimeout(scan.Config.ChannelRevealTimeout),
			channel.DBPath(scan.Config.ChannelDBPath),
			channel.SettleTimeout(scan.Config.ChannelSettleTimeout),
			channel.BlockDelay(scan.Config.BlockDelay),
			channel.IsClient(false),
		)
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
	// this.Channel.OverridePartners()
	return nil
}

func GetExternalIP(walletAddr string) (string, error) {
	log.Debugf("channel get hostinfo call %s", walletAddr)
	nodeAddr, err := storage.EDB.GetEndpoint(walletAddr)
	log.Debugf("%v, err: %v", nodeAddr, err)
	if err != nil {
		return "", err
	}
	if nodeAddr == nil {
		return "", errors.New("nodeAddr is nil")
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
	return this.Channel.CloseChannel(partnerAddr)
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

func (this *Node) MediaTransfer(paymentId int32, amount uint64, media string, to string) error {
	return this.Channel.MediaTransfer(paymentId, amount, media, to)
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
	var peer *storage.Endpoint
	peer, err := storage.EDB.GetEndpoint(partnerAddr)
	if err != nil {
		return "", err
	}
	if peer == nil {
		return "", errors.New("endpoint not registed")
	}
	return peer.NodeAddr.String(), nil
}

func (this *Node) channelExists(ci []*ch_actor.ChannelInfo, w string) bool {
	for _, ch := range ci {
		if ch.Address == w {
			return true
		}
	}
	return false
}

func (n *Node) GetFee() (uint64, error) {
	fee, err := n.Channel.GetFee()
	if err != nil {
		log.Errorf("GetFee err %v", err)
		return 0, err
	}
	return fee, nil
}

func (n *Node) SetFee(flat uint64) error {
	err := n.Channel.SetFee(chanCom.FeeAmount(flat))
	if err != nil {
		log.Errorf("SetFee err %v", err)
		return err
	}
	return nil
}