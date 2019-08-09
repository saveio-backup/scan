package channel

import (
	"container/list"
	"errors"
	"fmt"
	"time"

	"github.com/anacrolix/dht/krpc"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/dsp-go-sdk/channel"
	dspCfg "github.com/saveio/dsp-go-sdk/config"
	"github.com/saveio/dsp-go-sdk/store"
	chActorClient "github.com/saveio/pylons/actor/client"
	chanCom "github.com/saveio/pylons/common"
	"github.com/saveio/scan/common"
	"github.com/saveio/scan/common/config"
	"github.com/saveio/scan/storage"
	chain "github.com/saveio/themis-go-sdk"
	"github.com/saveio/themis/account"
	chainCom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
)

// GlbChannelSvr
var GlbChannelSvr *ChannelSvr

type ChannelSvr struct {
	Config  *dspCfg.DspConfig
	Chain   *chain.Chain
	Channel *channel.Channel
}

func NewChannelSvr(acc *account.Account, p2pActor *actor.PID) (*ChannelSvr, error) {
	cs := &ChannelSvr{}

	channelListenAddr := fmt.Sprintf("127.0.0.1:%d", int(config.Parameters.Base.PortBase+config.Parameters.Base.ChannelPortOffset))
	cs.Config = &dspCfg.DspConfig{
		DBPath:               config.DspDBPath(),
		ChainRpcAddr:         config.Parameters.Base.ChainRpcAddr,
		ChannelClientType:    config.Parameters.Base.ChannelClientType,
		ChannelListenAddr:    channelListenAddr,
		ChannelProtocol:      config.Parameters.Base.ChannelProtocol,
		ChannelRevealTimeout: config.Parameters.Base.ChannelRevealTimeout,
		ChannelDBPath:        config.ChannelDBPath(),
	}
	cs.Chain = chain.NewChain()
	cs.Chain.NewRpcClient().SetAddress([]string{cs.Config.ChainRpcAddr})
	if acc != nil {
		cs.Chain.SetDefaultAccount(acc)
	}

	var dbstore *store.LevelDBStore
	if len(cs.Config.DBPath) > 0 {
		var err error
		dbstore, err = store.NewLevelDBStore(cs.Config.DBPath)
		if err != nil {
			log.Errorf("init db err %s", err)
			return nil, nil
		}
	}

	if len(cs.Config.ChannelListenAddr) > 0 && acc != nil {
		var err error
		getHostCallBack := func(addr chainCom.Address) (string, error) {
			return GetExternalIP(addr.ToBase58())
		}

		cs.Channel, err = channel.NewChannelService(cs.Config, cs.Chain, getHostCallBack)
		if err != nil {
			log.Errorf("init channel err %s", err)
			return nil, err
		}
		chActorClient.SetP2pPid(p2pActor)
		if dbstore != nil {
			channelDB := store.NewChannelDB(dbstore)
			cs.Channel.SetChannelDB(channelDB)
		}
	}
	return cs, nil
}

func (this *ChannelSvr) Start() error {
	if this.Channel != nil {
		err := this.StartChannelService()
		if err != nil {
			return err
		}
	}
	return errors.New("ChannelSvr.channel is nil")
}

func (this *ChannelSvr) StartChannelService() error {
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
func (this *ChannelSvr) SetupPartnerHost(partners []string) {
	for _, addr := range partners {
		host, err := GetExternalIP(addr)
		if err != nil && host != "" {
			log.Infof("SetHostAddr partners, addr: %s host: %s", addr, host)
			this.Channel.SetHostAddr(addr, host)
		}
	}
}

// GetExternalIP. get external ip of wallet from dns nodes
func GetExternalIP(walletAddr string) (string, error) {
	w, _ := common.WHPTobyte(walletAddr, "")
	hpBytes, err := storage.TDB.Get(w)
	if err != nil {
		return "", err
	} else {
		var nodeAddr krpc.NodeAddr
		log.Infof("Channel.GetExternalIP wallAddr: %v, hpBytes: %v", walletAddr, hpBytes)
		nodeAddr.UnmarshalBinary(hpBytes)
		log.Infof("Channel.GetExternalIP nodeAddr: %s:%d", nodeAddr.IP, nodeAddr.Port)
		return fmt.Sprintf("%s://%s:%d", config.Parameters.Base.ChannelProtocol, nodeAddr.IP, nodeAddr.Port), nil
	}
}

//pylons api
func (this *ChannelSvr) OpenChannel(partnerAddr string, amount uint64) (chanCom.ChannelID, error) {
	return this.Channel.OpenChannel(partnerAddr, amount)
}

// func (this *ChannelSvr) CloseChannel(partnerAddr string) error {
// 	return this.Channel.ChannelClose(partnerAddr)
// }

func (this *ChannelSvr) QuerySpecialChannelDeposit(partnerAddr string) (uint64, error) {
	return this.Channel.GetTotalDepositBalance(partnerAddr)
}

func (this *ChannelSvr) DepositToChannel(partnerAddr string, totalDeposit uint64) error {
	return this.Channel.SetDeposit(partnerAddr, totalDeposit)
}

func (this *ChannelSvr) GetAllChannels() *channel.ChannelInfosResp {
	return this.Channel.AllChannels()
}

func (this *ChannelSvr) Transfer(paymentId int32, amount uint64, to string) error {
	return this.Channel.DirectTransfer(paymentId, amount, to)
}

func (this *ChannelSvr) GetChannelListByOwnerAddress(addr string, tokenAddr string) *list.List {
	//[TODO] call dsp-go-sdk function to return channel list
	//[NOTE] addr and token Addr should NOT be needed. addr mean PaymentNetworkID
	//tokenAddr mean TokenAddress. Need comfirm the behavior when integrate dsp-go-sdk with pylons
	return list.New()
}

func (this *ChannelSvr) ChannelWithdraw(partnerAddr string, amount uint64) error {
	success, err := this.Channel.Withdraw(partnerAddr, amount)
	if err != nil {
		return err
	}
	if !success {
		return errors.New("withdraw failed")
	}
	return nil
}

func (this *ChannelSvr) QueryHostInfo(partnerAddr string) (string, error) {
	return this.Channel.GetHostAddr(partnerAddr)
}

var startChannelHeight uint32

type FilterBlockProgress struct {
	Progress float32
	Start    uint32
	End      uint32
	Now      uint32
}

func (this *ChannelSvr) GetFilterBlockProgress() (*FilterBlockProgress, error) {
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
