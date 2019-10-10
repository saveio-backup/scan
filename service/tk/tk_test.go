package tk

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"

	tkpm "github.com/saveio/scan/p2p/actor/messages"
	"github.com/saveio/themis/common/log"
)

type TService struct {
	TMessageMap *sync.Map
}

type TMessageItem struct {
	MessageIdentifier *tkpm.MessageID
	Data              string
	ChStatus          chan AnnounceMessageStatus
}

func TestMessageMapChannels(t *testing.T) {
	ts := &TService{
		TMessageMap: new(sync.Map),
	}

	item := &TMessageItem{
		MessageIdentifier: &tkpm.MessageID{
			MessageId: uint64(GetMsgID()),
		},
		Data:     "hello",
		ChStatus: make(chan AnnounceMessageStatus),
	}

	ts.TMessageMap.Store(item.MessageIdentifier.MessageId, item)

	go PutToMessageChannel(item)
	go PutToMessageChannel(item)

	for {
		select {
		case status := <-item.ChStatus:
			msg, ok := ts.TMessageMap.Load(item.MessageIdentifier.MessageId)
			fmt.Printf("status %v, %v, %v\n", status, msg, ok)
			ts.TMessageMap.Delete(item.MessageIdentifier.MessageId)
		}
	}

	WaitToExit()
}

func PutToMessageChannel(item *TMessageItem) {
	item.ChStatus <- AnnounceMessageProcessed
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
