package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	pm "github.com/saveio/scan/p2p/actor/messages"
	tkAct "github.com/saveio/scan/p2p/actor/tracker/client"
	network "github.com/saveio/scan/p2p/networks/tracker"
	"github.com/saveio/scan/service/tk"
	"github.com/saveio/themis/common/log"
)

const (
	defaultMinTimeOut = 1500
	defaultMaxTimeOut = 60000
)

var TrackerServerPid *actor.PID

type MessageHandler func(msgData interface{}, pid *actor.PID)

type TrackerActorServer struct {
	net         *network.Network
	props       *actor.Props
	msgHandlers map[string]MessageHandler
	localPID    *actor.PID
	tkSrv       *tk.TrackerService
}

func NewTrackerActor(tkService *tk.TrackerService) (*TrackerActorServer, error) {
	var err error
	tkActorServer := &TrackerActorServer{
		msgHandlers: make(map[string]MessageHandler),
	}
	tkActorServer.props = actor.FromProducer(func() actor.Actor { return tkActorServer })
	tkActorServer.localPID, err = actor.SpawnNamed(tkActorServer.props, "scan_tracker_net_server")
	if err != nil {
		return nil, err
	}

	tkActorServer.tkSrv = tkService

	TrackerServerPid = tkActorServer.localPID
	return tkActorServer, nil
}

func (this *TrackerActorServer) SetNetwork(n *network.Network) {
	this.net = n
}

func (this *TrackerActorServer) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Restarting:
		log.Warn("[P2p]actor restarting")
	case *actor.Stopping:
		log.Warn("[P2p]actor stopping")
	case *actor.Stopped:
		log.Warn("[P2p]actor stopped")
	case *actor.Started:
		log.Debug("[P2p]actor started")
	case *actor.Restart:
		log.Warn("[P2p]actor restart")
	case *tkAct.ConnectReq:
		go func() {
			msg.Ret.Err = this.net.Connect(msg.Address)
			msg.Ret.Done <- true
		}()
	case *tkAct.CloseReq:
		go func() {
			msg.Ret.Err = this.net.Close(msg.Address)
			msg.Ret.Done <- true
		}()
	case *tkAct.SendReq:
		log.Infof("tkact SendReq %v, %s", msg.Data, msg.Address)
		go func() {
			msg.Ret.Err = this.net.Send(msg.Data, msg.Address)
			msg.Ret.Done <- true
		}()
	case *AnnounceReq:
		log.Infof("tkact AnnounceReq")
		go func() {
			annResp, err := this.tkSrv.HandleAnnounceRequestEvent(msg.Announce)
			if err == nil {
				msg.Ret.Ret = annResp
				msg.Ret.Err = nil
			} else {
				msg.Ret.Ret = annResp
				msg.Ret.Err = err
			}
			msg.Ret.Done <- true
		}()
	case *pm.AnnounceRequestMessage:
		log.Infof("tkact AnnounceRequestMessage %v", msg)
		go func() {
			this.tkSrv.ReceiveAnnounceMessage(msg, msg.GetRequest().From)
		}()
	case *pm.AnnounceResponseMessage:
		log.Infof("tkact AnnounceResponseMessage %v", msg.GetResponse().From)
		go func() {
			this.tkSrv.ReceiveAnnounceMessage(msg, "tcp://40.73.100.114:37835")
		}()
	default:
		log.Error("[P2PActor] receive unknown message type!")
	}
}

func (this *TrackerActorServer) RegMsgHandler(msgName string, handler MessageHandler) {
	this.msgHandlers[msgName] = handler
}

func (this *TrackerActorServer) UnRegMsgHandler(msgName string, handler MessageHandler) {
	delete(this.msgHandlers, msgName)
}

func (this *TrackerActorServer) GetLocalPID() *actor.PID {
	return this.localPID
}

func (this *TrackerActorServer) TKService() *tk.TrackerService {
	return this.tkSrv
}

func (this *TrackerActorServer) AnnounceRequestEndpointRegistry(req *pm.EndpointRegistryReq, targetDnsAddr string) (*pm.EndpointRegistryRet, error) {
	ret, err := this.AnnounceRequest(&pm.AnnounceRequest{
		EndpointRegistryReq: req,
		Event:               pm.AnnounceEvent_ENDPOINT_REGISTRY,
		Target:              targetDnsAddr,
	})
	if err != nil {
		return nil, err
	}
	return ret.EndpointRegistryRet, nil
}

func (this *TrackerActorServer) AnnounceRequestGetEndpointAddr(req *pm.QueryEndpointReq, targetDnsAddr string) (*pm.QueryEndpointRet, error) {
	ret, err := this.AnnounceRequest(&pm.AnnounceRequest{
		QueryEndpointReq: req,
		Event:            pm.AnnounceEvent_QUERY_ENDPOINT,
		Target:           targetDnsAddr,
	})
	log.Debugf("response %v %v", ret, err)
	if err != nil {
		return nil, err
	}
	return ret.QueryEndpointRet, nil
}

func (this *TrackerActorServer) AnnounceRequestCompleteTorrent(req *pm.CompleteTorrentReq, targetDnsAddr string) (*pm.CompleteTorrentRet, error) {
	ret, err := this.AnnounceRequest(&pm.AnnounceRequest{
		CompleteTorrentReq: req,
		Event:              pm.AnnounceEvent_COMPLETE_TORRENT,
		Target:             targetDnsAddr,
	})
	if err != nil {
		return nil, err
	}
	return ret.CompleteTorrentRet, nil
}

func (this *TrackerActorServer) AnnounceRequestTorrentPeers(req *pm.GetTorrentPeersReq, targetDnsAddr string) (*pm.GetTorrentPeersRet, error) {
	ret, err := this.AnnounceRequest(&pm.AnnounceRequest{
		GetTorrentPeersReq: req,
		Event:              pm.AnnounceEvent_QUERY_TORRENT_PEERS,
		Target:             targetDnsAddr,
	})
	if err != nil {
		return nil, err
	}
	return ret.GetTorrentPeersRet, nil
}

func (this *TrackerActorServer) AnnounceRequest(req *pm.AnnounceRequest) (*pm.AnnounceResponse, error) {
	ret := &AnnounceRet{
		Ret:  nil,
		Err:  nil,
		Done: make(chan bool, 1),
	}
	announceReq := &AnnounceReq{
		Announce: req,
		Ret:      ret,
	}
	this.GetLocalPID().Tell(announceReq)

	if err := waitForCallDone(announceReq.Ret.Done, "AnnounceRequest", defaultMaxTimeOut); err != nil {
		log.Debugf("AnnounceDone nil, err : %v", nil)
		return nil, err
	} else {
		log.Debugf("AnnounceDone %v, err: nil", announceReq.Ret.Ret)
		if announceReq.Ret.Ret.Timeout {
			return nil, errors.New("Timeout")
		}
		return announceReq.Ret.Ret, nil
	}
}

func waitForCallDone(c chan bool, funcName string, maxTimeOut int64) error {
	if maxTimeOut < defaultMinTimeOut {
		maxTimeOut = defaultMinTimeOut
	}
	select {
	case <-c:
		close(c)
		return nil
	case <-time.After(time.Duration(maxTimeOut) * time.Millisecond):
		return fmt.Errorf("function:[%s] timeout", funcName)
	}
}
