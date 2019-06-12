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

	chainRpcAddr := fmt.Sprintf("http://127.0.0.1:%d", config.DefaultConfig.CommonConfig.ChainRpcAddr)
	channelListenAddr := fmt.Sprintf("127.0.0.1:%d", config.DefaultConfig.ChannelConfig.ChannelPortOffset)

	cs.Config = &dspCfg.DspConfig{
		DBPath:               config.DefaultConfig.CommonConfig.DBPath,
		ChainRpcAddr:         chainRpcAddr,
		ChannelClientType:    config.DefaultConfig.ChannelConfig.ChannelClientType,
		ChannelListenAddr:    channelListenAddr,
		ChannelProtocol:      config.DefaultConfig.ChannelConfig.ChannelProtocol,
		ChannelRevealTimeout: config.DefaultConfig.ChannelConfig.ChannelRevealTimeout,
		ChannelDBPath:        config.DefaultConfig.ChannelConfig.ChannelDBPath,
	}
	cs.Chain = chain.NewChain()
	cs.Chain.NewRpcClient().SetAddress(cs.Config.ChainRpcAddr)
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
		cs.Channel, err = channel.NewChannelService(cs.Config, cs.Chain)
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
		host := this.GetExternalIP(addr)
		if host != "" {
			log.Infof("SetHostAddr partners, addr: %s host: %s", addr, host)
			this.Channel.SetHostAddr(addr, host)
		}
	}
}

// GetExternalIP. get external ip of wallet from dns nodes
func (this *ChannelSvr) GetExternalIP(walletAddr string) string {
	w, _ := common.WHPTobyte(walletAddr, "")
	hpBytes, err := storage.TDB.Get(w)
	var nodeAddr krpc.NodeAddr
	log.Infof("Channel.GetExternalIP wallAddr: %v, hpBytes: %v", walletAddr, hpBytes)
	nodeAddr.UnmarshalBinary(hpBytes)
	log.Infof("Channel.GetExternalIP nodeAddr: %s:%d", nodeAddr.IP, nodeAddr.Port)

	if err != nil {
		return ""
	} else {
		return fmt.Sprintf("%s://%s:%d", config.DefaultConfig.ChannelConfig.ChannelProtocol, nodeAddr.IP, nodeAddr.Port)
	}
}

//pylons api
func (this *ChannelSvr) OpenChannel(partnerAddr string) (chanCom.ChannelID, error) {
	return this.Channel.OpenChannel(partnerAddr)
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
