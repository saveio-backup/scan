package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/saveio/scan/common/config"
	"github.com/saveio/scan/tracker"
	trackercomm "github.com/saveio/scan/tracker/common"
	chainsdk "github.com/saveio/themis-go-sdk/utils"
	"github.com/saveio/themis/common/log"
)

func (this *Node) StartTrackerService() error {
	tkSvr := tracker.NewTKServer()
	tkSvr.Tsvr.SetPID(this.DnsNet.GetPID())
	go tkSvr.StartTrackerListening()
	log.Info("start tracker service success")
	return nil
}

func (this Node) RegOtherDnsEndpointsToSelf() error {
	publicAddr := this.ChannelNet.PublicAddr()
	index := strings.Index(publicAddr, "://")
	hostPort := publicAddr
	if index != -1 {
		hostPort = publicAddr[index+3:]
	}
	err := PutExternalIP(this.Account.Address.ToBase58(), hostPort)
	if err != nil {
		return err
	}

	ns, err := this.GetAllDnsNodes()
	if err != nil {
		return err
	}
	if len(ns) == 0 {
		log.Info("no dns nodes")
		return nil
	}
	for _, v := range ns {
		log.Debugf("DNS WalletAddr %s, IP: %v, port %v\n", v.WalletAddr.ToBase58(), string(v.IP), string(v.Port))
		err := PutExternalIP(v.WalletAddr.ToBase58(), fmt.Sprintf("%s:%s", string(v.IP), string(v.Port)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Node) RegSelfEndpointToOtherDns() error {
	endpointAddr := this.ChannelNet.PublicAddr()
	ns, err := this.Chain.Native.Dns.GetAllDnsNodes()
	if err != nil {
		return err
	}
	if len(ns) == 0 {
		log.Info("no dns nodes")
		return nil
	}

	trackerUrls := make([]string, 0)

	for _, v := range ns {
		if len(trackerUrls) >= MAX_DNS_CHANNELS_NUM_AUTO_OPEN_WITH {
			break
		}
		log.Debugf("DNS WalletAddr %s, IP: %v, port %v\n", v.WalletAddr.ToBase58(), string(v.IP), string(v.Port))
		trackerUrl := fmt.Sprintf("%s://%s:%d/announce", config.Parameters.Base.TrackerProtocol, v.IP, config.Parameters.Base.TrackerPortOffset)
		trackerUrls = append(trackerUrls, trackerUrl)
	}
	log.Debugf("trackerUrls %v\n", trackerUrls)
	if len(trackerUrls) == 0 {
		log.Info("no tracker urls")
		return nil
	}

	index := strings.Index(endpointAddr, "://")
	hostPort := endpointAddr
	if index != -1 {
		hostPort = endpointAddr[index+3:]
	}
	host, port, err := net.SplitHostPort(hostPort)
	log.Debugf("hostPort %v, host %v\n", hostPort, host)
	if err != nil {
		return err
	}
	netIp := net.ParseIP(host).To4()
	if netIp == nil {
		netIp = net.ParseIP(host).To16()
	}
	netPort, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return err
	}

	var wallet [20]byte
	copy(wallet[:], this.Account.Address[:])

	for _, trackerUrl := range trackerUrls {
		log.Debugf("trackerurl %s walletAddr: %v netIp:%v netPort:%v", trackerUrl, wallet, netIp, netPort)
		params := trackercomm.ApiParams{
			TrackerUrl: trackerUrl,
			Wallet:     wallet,
			IP:         netIp,
			Port:       uint16(netPort),
		}
		rawData, err := json.Marshal(params)
		if err != nil {
			return err
		}

		sigData, err := chainsdk.Sign(this.CurrentAccount(), rawData)
		if err != nil {
			return err
		}
		request := func(resp chan *trackerResp) {
			log.Debugf("start RegEndPoint %s ipport %v:%v", trackerUrl, netIp, netPort)
			err := tracker.RegEndPoint(trackerUrl, sigData, this.CurrentAccount().PublicKey, wallet, netIp, uint16(netPort))
			if err != nil {
				log.Errorf("req endpoint failed, err %s", err)
			}
			resp <- &trackerResp{
				ret: nil,
				err: err,
			}
		}
		_, err = this.trackerReq(trackerUrl, request)

		if err != nil {
			continue
		}
	}

	return nil
}

type trackerResp struct {
	ret interface{}
	err error
}

func (this *Node) trackerReq(trackerUrl string, request func(chan *trackerResp)) (interface{}, error) {
	done := make(chan *trackerResp, 1)
	go request(done)
	for {
		select {
		case ret := <-done:
			return ret.ret, ret.err
		case <-time.After(TRACKER_SERVICE_TIMEOUT):
			log.Errorf("tracker: %s request timeout", trackerUrl)
			return nil, errors.New("tracker request timeout")
		}
	}
}
