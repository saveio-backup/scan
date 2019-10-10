package server

import (
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	pm "github.com/saveio/scan/p2p/actor/messages"
	tkAct "github.com/saveio/scan/p2p/actor/tracker/client"
	network "github.com/saveio/scan/p2p/networks/tracker"
	"github.com/saveio/scan/service/tk"
	"github.com/saveio/themis/common/log"
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
		log.Infof("tkact SendReq %v", msg.Data, msg.Address)
		go func() {
			msg.Ret.Err = this.net.Send(msg.Data, msg.Address)
			msg.Ret.Done <- true
		}()
	case *pm.AnnounceRequest:
		log.Infof("tkact AnnounceRequest %v", msg)
		annResp, err := this.tkSrv.HandleAnnounceRequestEvent(msg)
		if err != nil {
			ctx.Sender().Request(annResp, ctx.Self())
		} else {
			ctx.Sender().Request(err, ctx.Self())
		}
	case *pm.AnnounceRequestMessage:
		log.Infof("tkact AnnounceRequestMessage %v", msg)
		this.tkSrv.ReceiveAnnounceMessage(msg, ctx.Self().String())
	case *pm.AnnounceResponseMessage:
		this.tkSrv.ReceiveAnnounceMessage(msg, ctx.Self().String())
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

func (this *TrackerActorServer) AnnounceRequest(req *pm.AnnounceRequest) (*pm.AnnounceResponse, error) {
	log.Debugf("AnnounceRequest")
	future := this.GetLocalPID().RequestFuture(req, 30*time.Second)
	if ret, err := future.Result(); err != nil {
		log.Error("", err)
		return nil, err
	} else {
		return ret.(*pm.AnnounceResponse), nil
	}
}
